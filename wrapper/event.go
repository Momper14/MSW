package wrapper

import (
	"fmt"
)

// Event enum of events
type Event int

// posible events
const (
	EmptyEvent Event = iota + 1
	StartedEvent
	StoppedEvent
	StartEvent
	StopEvent
)

var eventmap = map[Event]string{
	EmptyEvent:   "empty",
	StartedEvent: "started",
	StoppedEvent: "stopped",
	StartEvent:   "start",
	StopEvent:    "stop",
}

func (e Event) String() string {
	if val, ok := eventmap[e]; ok {
		return val
	}

	return "unknown"
}

// EventFor returns Event for the given string
// ignores errors
func EventFor(e string) Event {
	event, _ := EventForE(e)
	return event
}

// EventForE returns Event for the given string
func EventForE(e string) (Event, error) {
	for k, v := range eventmap {
		if v == e {
			return k, nil
		}
	}

	return Event(0), fmt.Errorf("no known state for %s", e)
}

// Validate validates that the value is a valide enum value
func (e Event) Validate() (ok bool) {
	_, ok = eventmap[e]
	return
}
