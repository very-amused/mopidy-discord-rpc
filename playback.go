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

	// Each time a goroutine may want to be canceled early, it is added to []Cancel
	// Iterating over the range here, and sending a message to each channel will shutdown each goroutine and give the current one authoritve control of state
	Cancel *chan (bool)
	Done   chan (bool)
}

func (p *Playback) init() {
	// Create the playback's Done channel
	p.Done = make(chan bool)
	// Initialize and stop the playback ticker
	p.Ticker = time.NewTicker(time.Second)
	p.Ticker.Stop()
	go func() {
		for {
			select {
			// Do not destroy and recreate the ticker, only stop and start it on relevant play/pause events
			case <-p.Ticker.C:
				p.write()
				p.Elapsed += time.Second

			// p.Done closes the playback goroutine
			case <-p.Done:
				p.Ticker.Stop()
				return
			}
		}
	}()
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

func (p *Playback) pause() {
	p.Ticker.Stop()
	p.IsPlaying = false
	p.write()
}

func (p *Playback) play() {
	// Cancel any pending syncs
	p.IsPlaying = true
	p.write()
	p.Ticker.Reset(time.Second)
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

	if offset.Milliseconds() == 0 {
		p.play()
		return
	}
	// Create and start a timer that will go off at the next second
	start := time.Now()
	timer := time.NewTimer(time.Second - offset)

	// Block until the timer is done, return if the track was paused in this time
	cancel := make(chan bool)
	p.Cancel = &cancel
	go func() {
		for {
			select {
			case <-timer.C:
				p.Elapsed += time.Now().Sub(start)
				p.play()

			case <-cancel:
				timer.Stop()
				return
			}
		}
	}()
}

var playback = Playback{}

func init() {
	playback.init()
}
