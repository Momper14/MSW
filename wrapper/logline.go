package wrapper

import (
	"fmt"
	"regexp"
)

// logRegex Regexp to parse Minecraft Server Logs
var logRegex = regexp.MustCompile(`\[([0-9:]*)\] \[([A-z(-| )#0-9]*)\/([A-z #]*)\] \[(.*)\]: (.*)`)

// eventToRegexpMap maps events to the coressponding log line
var eventToRegexpMap = map[Event]*regexp.Regexp{
	StartedEvent: regexp.MustCompile(`Done (?s)(.*)! For help, type "help"`),
	StartEvent:   regexp.MustCompile(`Starting minecraft server version (.*)`),
	StopEvent:    regexp.MustCompile(`Stopping (.*) server`),
}

// logLine parts of the minecraft server logs
type logLine struct {
	timestamp  string
	threadName string
	level      string
	mod        string
	output     string
}

// parseToLogLine parses a string to logline
func parseToLogLine(line string) (*logLine, error) {
	matches := logRegex.FindAllStringSubmatch(line, 4)

	if matches == nil || len(matches[0]) < 5 {
		return nil, fmt.Errorf("unknown logline format for: %s", line)
	}

	return &logLine{
		timestamp:  matches[0][1],
		threadName: matches[0][2],
		level:      matches[0][3],
		mod:        matches[0][4],
		output:     matches[0][5],
	}, nil
}

// toEvent gets the type of event for the logline
func (ll *logLine) toEvent() Event {
	for e, r := range eventToRegexpMap {
		if r.Match([]byte(ll.output)) {
			return e
		}
	}
	return EmptyEvent
}
