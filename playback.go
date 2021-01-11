package main

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/very-amused/mopidy-discord-rpc/discord"
)

// Playback state
type Playback struct {
	Title     string
	Artists   string
	IsPlaying bool
	Source    string
	Elapsed   time.Duration
	Total     time.Duration
	Ticker    *time.Ticker
	Done      chan (bool)
}

func (p *Playback) write() {
	var playbackState string
	if p.IsPlaying {
		playbackState = "Playing"
	} else {
		playbackState = "Paused"
	}
	elapsedMinutes := int(math.Floor(p.Elapsed.Minutes()))
	elapsedSeconds := int(math.Floor(math.Mod(p.Elapsed.Seconds(), 60)))
	totalMinutes := int(math.Floor(p.Total.Minutes()))
	totalSeconds := int(math.Floor(math.Mod(p.Total.Seconds(), 60)))

	discord.Presence.Details = fmt.Sprintf("%s - %s", p.Artists, p.Title)
	discord.Presence.State = fmt.Sprintf("%s (%02d:%02d/%02d:%02d)", playbackState, elapsedMinutes, elapsedSeconds, totalMinutes, totalSeconds)
	if p.Source == "spotify" {
		discord.Presence.SmallImageKey = "spotify"
	} else {
		discord.Presence.SmallImageKey = ""
	}
	discord.UpdateRPC()
}

var playback = Playback{
	Done: make(chan bool)}

func (p *Playback) pause() {
	// If already playing, close the existing ticker goroutine (IsPlaying guarantees that a goroutine is listening on the done channe)
	if p.IsPlaying {
		p.Done <- true
	}
	p.IsPlaying = false
	p.write()
}

func (p *Playback) play() {
	p.Ticker = time.NewTicker(time.Second)
	p.IsPlaying = true

	p.write()

	go func() {
		for {
			select {
			case <-p.Ticker.C:
				p.write()
				p.Elapsed += time.Second

			case <-p.Done:
				p.Ticker.Stop()
				return
			}
		}
	}()
}

// setDetails - Set playback details (artists and title)
func (p *Playback) setDetails(track MopidyTrack) {
	// Separate artist names with commas
	p.Artists = ""
	for i, artist := range track.Artists {
		p.Artists += artist.Name
		if i < len(track.Artists)-1 {
			p.Artists += ", "
		}
	}
	if strings.HasPrefix(track.URI, "spotify") {
		p.Source = "spotify"
	} else {
		p.Source = "local"
	}
	p.Title = track.Name
}

// Sync the playback ticker to Mopidy's ticker, resuming playback when in sync
func (p *Playback) syncAndPlay(elapsed time.Duration) {
	// Get offset from the previous second
	p.Elapsed = elapsed
	offset := time.Duration(math.Floor(
		math.Mod(float64(elapsed.Milliseconds()), 1000))) * time.Millisecond

	// Create and start a timer that will go off at the nearest second
	start := time.Now()
	var timer *time.Timer
	if offset.Milliseconds() == 0 {
		timer = time.NewTimer(0)
	} else {
		timer = time.NewTimer(time.Second - offset)
	}

	// Listen for p.Done here, which would fire if the user paused the track before the nearest second was reached
	// To synchronize perfectly with the Mopidy tick (millisecond precision),
	// add the exact amount of time elapsed since the timer was initialized when a channel responds
	select {
	case <-p.Done:
		p.Elapsed += time.Now().Sub(start)
		timer.Stop()
		return

	case <-timer.C:
		p.Elapsed += time.Now().Sub(start)
		p.play()
	}
}
