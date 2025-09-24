package main

import (
	"encoding/json"
	"fmt"

	"github.com/SUSE/connect-ng/pkg/registration"
	"github.com/rtamalin/rmt-client-testing/internal/clientstore"
)

type SysInfo registration.SystemInformation

func (si *SysInfo) Load(fileId clientstore.FileId, clientStore *clientstore.ClientStore) (err error) {
	fileType := clientstore.SYS_INFO_TYPE
	siBytes, err := clientStore.ReadFile(fileId, fileType)
	if err != nil {
		err = fmt.Errorf(
			"failed to read system information from %q: %w",
			fileId.Path(fileType),
			err,
		)
		return
	}

	err = json.Unmarshal(siBytes, si)
	if err != nil {
		err = fmt.Errorf(
			"failed to unmarshal system information JSON: %w",
			err,
		)
		return
	}

	return
}

func (si *SysInfo) Save(fileId clientstore.FileId, clientStore *clientstore.ClientStore) (err error) {
	siBytes, err := json.Marshal(si)
	if err != nil {
		err = fmt.Errorf(
			"failed to marshal system information JSON: %w",
			err,
		)
		return
	}

	fileType := clientstore.SYS_INFO_TYPE
	err = clientStore.WriteFile(fileId, fileType, []byte(siBytes), 0o644)
	if err != nil {
		err = fmt.Errorf(
			"failed to read system information from %q: %w",
			fileId.Path(fileType),
			err,
		)
		return
	}

	return
}
