package model

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// MessageType enum of message types
type MessageType int

// possible types
const (
	TypeLog MessageType = iota + 1
	TypeError
	TypeState
)

var typeToString = map[MessageType]string{
	TypeLog:   "LOG",
	TypeError: "ERROR",
	TypeState: "STATE",
}

var typeForString = map[string]MessageType{
	"LOG":   TypeLog,
	"ERROR": TypeError,
	"STATE": TypeState,
}

func (t MessageType) String() string {
	if val, ok := typeToString[t]; ok {
		return val
	}

	return "unknown"
}

// TypeFor returns Type for the given string
// ignores errors
func TypeFor(s string) MessageType {
	t, _ := TypeForE(s)
	return t
}

// TypeForE returns Type for the given string
func TypeForE(s string) (MessageType, error) {
	if val, ok := typeForString[s]; ok {
		return val, nil
	}

	return MessageType(0), fmt.Errorf("no known type for %s", s)
}

// Validate validates that the value is a valide enum value
func (t MessageType) Validate() (ok bool) {
	_, ok = typeToString[t]
	return
}

// MarshalJSON marshals the enum as a quoted json string
func (t MessageType) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(typeToString[t])
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

// UnmarshalJSON unmashals a quoted json string to the enum value
func (t *MessageType) UnmarshalJSON(b []byte) error {
	var j string
	err := json.Unmarshal(b, &j)
	if err != nil {
		return err
	}
	*t = typeForString[j]
	return nil
}
