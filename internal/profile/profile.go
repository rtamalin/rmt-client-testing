package profile

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
)

type ProfileInfo struct {
	ProfileID   string `json:"profileId"`
	ProfileData any    `json:"profileData"`
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

	pi.ProfileID = hex.EncodeToString(hasher.Sum(nil))
	pi.ProfileData = data

}

func NewProfileInfo(data any) *ProfileInfo {
	pi := new(ProfileInfo)
	pi.Init(data)

	return pi
}
