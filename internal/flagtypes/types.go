package flagtypes

import (
	"fmt"
	"strconv"
)

// Uint32 is a uint32 compatible type for use with the flag module
type Uint32 uint32

func (u *Uint32) String() string {
	return strconv.FormatUint(uint64(*u), 10)
}

func (u *Uint32) Set(v string) (err error) {
	val, err := strconv.ParseUint(v, 10, 32)
	if err == nil {
		*u = Uint32(val)
		return
	}

	err = fmt.Errorf(
		"failed to parse %q as a Uint32: %w",
		v,
		err,
	)
	return
}
