package main

import (
	"encoding/json"
	"math/rand"
	"strconv"
	"time"
)

// Initialize RPC status on start
func initRPC() {
	// Identify communication with a random id
	id := strconv.Itoa(rand.Int())

	var playbackState struct {
		MopidyRPCResponse
		Result string `json:"result"`
	}
	makeRPCRequest(id, "core.playback.get_state")
	readRPCResponse(id, &playbackState)
	if playbackState.Result == "stopped" {
		return
	}

	// Get track info
	var trackInfo struct {
		MopidyRPCResponse
		Result MopidyTLTrack `json:"result"`
	}
	makeRPCRequest(id, "core.playback.get_current_tl_track")
	readRPCResponse(id, &trackInfo)
	playback.setDetails(trackInfo.Result.Track)
	playback.Total = trackInfo.Result.Track.Length * time.Millisecond
	playback.write()

	var timeInfo struct {
		MopidyRPCResponse
		Result time.Duration `json:"result"`
	}
	makeRPCRequest(id, "core.playback.get_time_position")
	readRPCResponse(id, &timeInfo)
	if playbackState.Result == "playing" {
		playback.syncAndPlay(timeInfo.Result*time.Millisecond + time.Second)
	} else {
		playback.Elapsed = timeInfo.Result * time.Millisecond
		playback.write()
	}
}

func makeRPCRequest(id string, method string) {
	conn.WriteJSON(MopidyRPCRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method})
}

func readRPCResponse(id string, dest interface{}) {
	var body []byte
	var info MopidyRPCResponse

	// Read messages until a JSON-RPC message matching the given id is found, then decode it into dest
	for {
		_, body, _ = conn.ReadMessage()
		json.Unmarshal(body, &info)
		if info.JSONRPC == "2.0" && info.ID == id {
			json.Unmarshal(body, dest)
			break
		}
	}
}
