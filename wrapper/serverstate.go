package wrapper

import "fmt"

// ServerState enum of server states
type ServerState int

// possible states
const (
	ServerOffline ServerState = iota + 1
	ServerOnline
	ServerStarting
	ServerStopping
)

var statemap = map[ServerState]string{
	ServerOffline:  "offline",
	ServerOnline:   "online",
	ServerStarting: "starting",
	ServerStopping: "stopping",
}

func (s ServerState) String() string {
	if val, ok := statemap[s]; ok {
		return val
	}

	return "unknown"
}

// ServerStateFor returns State for the given string
// ignores errors
func ServerStateFor(s string) ServerState {
	state, _ := ServerStateForE(s)
	return state
}

// ServerStateForE returns State for the given string
func ServerStateForE(s string) (ServerState, error) {
	for k, v := range statemap {
		if v == s {
			return k, nil
		}
	}

	return ServerState(0), fmt.Errorf("no known state for %s", s)
}

// Validate validates that the value is a valide enum value
func (s ServerState) Validate() (ok bool) {
	_, ok = statemap[s]
	return
}
