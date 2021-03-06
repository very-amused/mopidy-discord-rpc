package main

import (
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

	TimePosition *time.Duration `json:"time_position"`
}

// MopidyTrack - Information about a track
type MopidyTrack struct {
	URI     string         `json:"uri"`
	Name    string         `json:"name"`
	Artists []MopidyArtist `json:"artists"`
	Length  time.Duration  `json:"length"`
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

// #region Mopidy JSON-RPC 2.0 communication

// MopidyRPCRequest - Request for RPC information from mopidy
type MopidyRPCRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      string `json:"id"`
	Method  string `json:"method"`
}

type MopidyRPCResponse struct {
	JSONRPC string `json:"jsonrpc"`
	ID      string `json:"id"`
}

// #endregion

// mopidyEvent - Event message fired from mopidy
type mopidyEvent string

// Declaration of all Mopidy event messages listened to
const (
	// Track started playing, details in Track
	trackPlaybackStarted mopidyEvent = "track_playback_started"
	// Track was resumed
	trackPlaybackResumed mopidyEvent = "track_playback_resumed"
	// Playback state changed
	playbackStateChanged mopidyEvent = "playback_state_changed"

	seeked mopidyEvent = "seeked"
)

var (
	url string
)

var dialer = websocket.Dialer{
	HandshakeTimeout: 2 * time.Second}

// Handle RPC event messages
func onMessage(message MopidyRPCMessage) {
	if message.Event == nil {
		return
	}

	// If an event has started a goroutine, tell it to exit early
	if playback.Cancel != nil {
		deadlock := time.NewTimer(5 * time.Second)
		select {
		case *playback.Cancel <- true:
			deadlock.Stop()
			break
		case <-deadlock.C:
			panic("playback.Cancel has blocked for 5 seconds, deadlock detected!")
		}
		playback.Cancel = nil
	}

	switch mopidyEvent(*message.Event) {
	case trackPlaybackStarted:
		playback.Elapsed = 0
		playback.Total = (*message.TLTrack).Track.Length * time.Millisecond
		playback.setDetails((*message.TLTrack).Track)
		break

	case trackPlaybackResumed:
		// Write that the track is playing without delay
		playback.IsPlaying = true
		playback.Total = (*message.TLTrack).Track.Length * time.Millisecond
		playback.Elapsed = *message.TimePosition * time.Millisecond
		playback.setDetails((*message.TLTrack).Track)
		break

	case playbackStateChanged:
		switch *message.NewState {
		case "stopped":
			playback.clear()
			break
		case "playing":
			playback.play()
			break
		case "paused":
			playback.pause()
		}
		break

	case seeked:
		playback.Elapsed = *message.TimePosition * time.Millisecond
		playback.play()
		break
	}
}
