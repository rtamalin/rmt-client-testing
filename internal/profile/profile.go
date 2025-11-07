package profile

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
)

type ProfileInfo struct {
	Digest string `json:"digest"`
	Data   any    `json:"data"`
}

func (pi *ProfileInfo) Init(data any) {
	piBytes, err := json.Marshal(data)
	if err != nil {
		log.Fatalf(
			"Failed to generate profile info JSON for %v: %s",
			data,
			err.Error(),
		)
	}

	hasher := sha256.New()
	hasher.Write(piBytes)

	pi.Digest = hex.EncodeToString(hasher.Sum(nil))
	pi.Data = data

}

func NewProfileInfo(data any) *ProfileInfo {
	pi := new(ProfileInfo)
	pi.Init(data)

	return pi
}
