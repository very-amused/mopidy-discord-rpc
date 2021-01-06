package main

import (
	"flag"
)

func main() {
	// Allow the user to specify a custom URL
	flag.StringVar(&url, "url", "ws://localhost:6680/mopidy/ws", "Websocket URL (including port) used to connect to Mopidy.")
	flag.Parse()

	// Connect to mopidy websocket, 2s timeout
	conn, _, err := dialer.Dial(url, nil)
	if err != nil {
		panic(err)
	}

	defer conn.Close()
	for {
		var message MopidyRPCMessage
		conn.ReadJSON(&message)
		onMessage(message)
	}
}
