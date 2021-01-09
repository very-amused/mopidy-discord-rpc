package discord

/*
#cgo CFLAGS: -I/usr/include
#cgo LDFLAGS: -L/usr/lib -ldiscord-rpc
#include <discord_register.h>
#include <discord_rpc.h>
*/
import "C"

var presence = C.struct_DiscordRichPresence{}

// Presence - RPC struct mapped to C types and passed to discord when UpdateRPC is called
var Presence = struct {
	Details       string
	State         string
	LargeImageKey string
}{}

// InitRPC - Initialize connection to Discord and begin updating RPC
func InitRPC(clientID string) {
	C.Discord_Initialize(C.CString(clientID), nil, 0, nil)
	UpdateRPC()
}

// ShutdownRPC - Shutdown connection to Discord
func ShutdownRPC() {
	C.Discord_Shutdown()
}

// UpdateRPC - Update the RPC presence to whatever is currently set
func UpdateRPC() {
	presence.details = C.CString(Presence.Details)
	presence.state = C.CString(Presence.State)
	presence.largeImageKey = C.CString(Presence.LargeImageKey)

	C.Discord_UpdatePresence(&presence)
}
