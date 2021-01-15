package main

import (
	"flag"

	discord "github.com/very-amused/mopidy-discord-rpc/discord"
)

const clientID = "796397801797320715"
const largeImageKey = "logo"

func main() {
	// Allow the user to specify a custom URL
	flag.StringVar(&url, "url", "ws://localhost:6680/mopidy/ws", "Websocket URL (including port) used to connect to Mopidy.")
	flag.Parse()

	// Initialize discord RPC
	discord.InitRPC(clientID)
	discord.Presence.LargeImageKey = largeImageKey
	defer discord.ShutdownRPC()

	// Connect to mopidy websocket, 2s timeout
	conn, _, err := dialer.Dial(url, nil)
	if err != nil {
		panic(err)
	}

	defer conn.Close()
	defer func() {
		playback.Done <- true
	}()

	for {
		var message MopidyRPCMessage
		conn.ReadJSON(&message)
		onMessage(message)
	}
}
