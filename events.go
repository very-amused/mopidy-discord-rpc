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

// Mopidy events enum
const (
	// New playback state, compare NewState with OldState before updating
	playbackStateChanged = "playback_state_changed"
	// Track started playing, details in Track
	trackPlaybackStarted = "track_playback_started"
	// Track stopped playing
	trackPlaybackPaused = "track_playback_paused"
	// Track was resumed
	trackPlaybackResumed = "track_playback_resumed"
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

	switch *message.Event {
	case trackPlaybackStarted:
		playback.Elapsed = 0
		playback.Total = (*message.TLTrack).Track.Length * time.Millisecond
		playback.setDetails((*message.TLTrack).Track)
		playback.play()

	case trackPlaybackResumed:
		// Write that the track is playing without delay
		playback.IsPlaying = true
		playback.Total = (*message.TLTrack).Track.Length * time.Millisecond
		playback.setDetails((*message.TLTrack).Track)
		playback.write()
		// Sync the ticker and play
		playback.syncAndPlay((*message.TimePosition) * time.Millisecond)

	case trackPlaybackPaused:
		playback.pause()
	}
}
