package pusher

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type Pusher struct {
	Key, Secret  string
	App_id       int64
	usingHttps   bool
	auth_version string
}

type PusherRequest struct {
	Evnet    string   `json:"name"`
	Channels []string `json:"channels"`
	Data     string   `json:"data"`
}

type PusherChannelResponse struct {
	Channels map[string]PusherChannel `json:"channels"`
}

type PusherChannel struct {
	UserCount int `json:"user_count"`
}

func (self *Pusher) SetHttps(flag bool) {
	self.usingHttps = flag
}

func (self Pusher) GetChannels() ([]string, error) {
	var result PusherChannelResponse
	requestUrl := self.get_Request()
	requestUrl.Path = fmt.Sprintf("/apps/%d/channels", self.App_id)

	self.signing_RequestURL(requestUrl)

	resp, err := http.Get(requestUrl.String())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Http Get Error:", err)
		return nil, err
	}
	log.Printf("Status Code: %+v", resp.StatusCode)

	contents, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	log.Printf("Response: %+v", string(contents))

	if err := json.Unmarshal(contents, &result); err != nil {
		fmt.Fprintf(os.Stderr, "JSON Unmarshal Error:", err)
		return nil, err
	}

	keys := make([]string, len(result.Channels))
	i := 0
	for name, _ := range result.Channels {
		keys[i] = name
		i = i + 1
	}
	return keys, nil
}

func (self Pusher) Trigger(channels []string, event string, message interface{}) error {

	requestUrl := self.get_Request()
	requestUrl.Path = fmt.Sprintf("/apps/%d/events", self.App_id)

	encoded_message, err := json.Marshal(message)
	if err != nil {
		fmt.Fprintf(os.Stderr, "JSON Encoding Error:", err)
		return err
	}

	request_body, err := json.Marshal(
		&PusherRequest{event, channels, string(encoded_message)})

	if err != nil {
		fmt.Fprintf(os.Stderr, "JSON Encoding Error:", err)
		return err
	}

	self.signing_Body(requestUrl, request_body)
	self.signing_RequestURL(requestUrl)

	resp, err := http.Post(requestUrl.String(),
		"application/json",
		bytes.NewReader(request_body))
	if err != nil {
		return err
	}
	log.Printf("Status Code: %+v", resp.StatusCode)

	contents, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	log.Printf("Response: %+v", string(contents))

	return nil
}

func (self Pusher) get_Request() *url.URL {
	requestUrl := new(url.URL)
	if self.usingHttps {
		requestUrl.Scheme = "https"
	} else {
		requestUrl.Scheme = "http"
	}
	requestUrl.Host = "api.pusherapp.com"
	q := url.Values{}
	q.Add("auth_key", self.Key)

	timestamp := fmt.Sprintf("%.0f", time.Since(
		time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC)).Seconds())

	q.Add("auth_timestamp", timestamp)
	q.Add("auth_version", "1.0")
	requestUrl.RawQuery = q.Encode()
	return requestUrl
}

func (self Pusher) signing_Body(requestUrl *url.URL, request_body []byte) {
	q := requestUrl.Query()
	q.Add("body_md5", fmt.Sprintf("%x", md5.Sum(request_body)))
	requestUrl.RawQuery = q.Encode()
}

func (self Pusher) signing_RequestURL(requestUrl *url.URL) {

	var method string
	if _, ok := requestUrl.Query()["body_md5"]; ok {
		method = "POST"
	} else {
		method = "GET"
	}
	source := strings.Join([]string{method, requestUrl.Path, requestUrl.RawQuery}, "\n")

	q := requestUrl.Query()
	mac := hmac.New(sha256.New, []byte(self.Secret))
	mac.Write([]byte(source))
	signature := hex.EncodeToString(mac.Sum(nil))

	q.Add("auth_signature", signature)
	requestUrl.RawQuery = q.Encode()
}
