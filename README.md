pusher
======

The lightweight Pusher library for Go language

Usage
-----

```Go
package main

import (
	"fmt"

	"github.com/tomahawk28/pusher"
)

type SampleMsg struct {
	Title   string `json:"title"`
	Message string `json:"message"`
}

func main() {
	pu := pusher.Pusher{
		Key:    "ef0affaffaffc3e8b5",
		Secret: "a99e8f18374691561d",
		App_id: 99999,
	}

	pu.SetHttps(true)
	pu.Trigger([]string{"test_channel"}, "my_event", &SampleMsg{"Greeting", "Thank you for waiting"})

	if ch, err := pu.GetChannels(); err != nil {
		panic("Die!")
	} else {
		for name, _ := range ch.Channels {
			fmt.Println("Channel name:", name)
		}
	}

}
```
