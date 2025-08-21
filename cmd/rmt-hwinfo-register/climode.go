package main

import (
	"fmt"
	"strings"
)

// CliAction is the type used to specify the mode of operation
type CliAction uint

const (
	ACTION_REGISTER CliAction = iota
	ACTION_UPDATE
	ACTION_DEREGISTER
	numActions
)

var modeNames = [numActions]string{
	ACTION_REGISTER:   "register",
	ACTION_UPDATE:     "update",
	ACTION_DEREGISTER: "deregister",
}

func (m *CliAction) String() (mode string) {
	if *m < numActions {
		mode = modeNames[*m]
	} else {
		mode = "UNKNOWN_CLI_ACTION"
	}
	return
}

func (m *CliAction) Set(value string) (err error) {
	checkValue := strings.ToLower(value)
	for i := CliAction(0); i < numActions; i++ {
		if checkValue == modeNames[i] {
			*m = i
			return
		}
	}

	err = fmt.Errorf(
		"invalid CLI mode %q specified, must be one of: %s",
		value,
		strings.Join(modeNames[:], ","),
	)

	return
}
