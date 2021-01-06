package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

// MopidyRPCMessage - An event message from the Mopidy RPC
type MopidyRPCMessage struct {
	Event *string `json:"event"`

	// State changes
	OldState *string `json:"old_state"`
	NewState *string `json:"new_state"`

	// Track (TL context)
	TLTrack *MopidyTLTrack `json:"tl_track"`
}

// MopidyTrack - Information about a track
type MopidyTrack struct {
	URI     string         `json:"uri"`
	Name    string         `json:"name"`
	Artists []MopidyArtist `json:"artists"`
}

// MopidyTLTrack - Information about a track (in the context of the tracklist)
type MopidyTLTrack struct {
	TrackNo uint        `json:"track_no"`
	Track   MopidyTrack `json:"track"`
}

// MopidyArtist - Information about an Artist
type MopidyArtist struct {
	URI  string `json:"uri"`
	Name string `json:"name"`
}

// Mopidy events enum
const (
	// New playback state, compare NewState with OldState before updating
	playbackStateChanged = "playback_state_changed"
	// Track started playing, details in Track
	trackPlaybackStarted = "track_playback_started"
)

var (
	url           string
	playbackState = "paused"
)

var dialer = websocket.Dialer{
	HandshakeTimeout: 2 * time.Second}

// Handle RPC event messages
func onMessage(message MopidyRPCMessage) {
	if message.Event == nil {
		return
	}

	switch *message.Event {
	case playbackStateChanged:
		if *message.NewState != *message.OldState {
			playbackState = *message.NewState
			fmt.Println(playbackState)
		}

	case trackPlaybackStarted:
		track := (*message.TLTrack).Track
		var formattedArtists strings.Builder
		for _, artist := range track.Artists {
			formattedArtists.WriteString(artist.Name)
		}
		fmt.Printf("Now playing:\nName: %s\nArtists: %s\n", track.Name, formattedArtists.String())
	}
}
