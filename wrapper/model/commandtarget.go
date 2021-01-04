package model

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// CommandTarget enum of message targets
type CommandTarget int

// possible targets
const (
	TargetWrapper CommandTarget = iota + 1
	TargetServer
)

var targetToString = map[CommandTarget]string{
	TargetWrapper: "WRAPPER",
	TargetServer:  "SERVER",
}

var targetForString = map[string]CommandTarget{
	"WRAPPER": TargetWrapper,
	"SERVER":  TargetServer,
}

func (t CommandTarget) String() string {
	if val, ok := targetToString[t]; ok {
		return val
	}

	return "unknown"
}

// TargetFor returns Target for the given string
// ignores errors
func TargetFor(s string) CommandTarget {
	t, _ := TargetForE(s)
	return t
}

// TargetForE returns Target for the given string
func TargetForE(s string) (CommandTarget, error) {
	if val, ok := targetForString[s]; ok {
		return val, nil
	}

	return CommandTarget(0), fmt.Errorf("no known target for %s", s)
}

// Validate validates that the value is a valide enum value
func (t CommandTarget) Validate() (ok bool) {
	_, ok = targetToString[t]
	return
}

// MarshalJSON marshals the enum as a quoted json string
func (t CommandTarget) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(targetToString[t])
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

// UnmarshalJSON unmashals a quoted json string to the enum value
func (t *CommandTarget) UnmarshalJSON(b []byte) error {
	var j string
	err := json.Unmarshal(b, &j)
	if err != nil {
		return err
	}
	*t = targetForString[j]
	return nil
}
