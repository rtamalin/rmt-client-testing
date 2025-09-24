package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/rtamalin/rmt-client-testing/internal/clientstore"
)

type RegInfo struct {
	SccCreds SccCredentials `json:"scc_creds"`
}

func RegInfoExists(fileId clientstore.FileId, clientStore *clientstore.ClientStore) bool {
	return clientStore.Exists(fileId, clientstore.REG_INFO_TYPE)
}

func (ri *RegInfo) Delete(fileId clientstore.FileId, clientStore *clientstore.ClientStore) (err error) {
	fileType := clientstore.REG_INFO_TYPE
	err = clientStore.Delete(fileId, fileType)
	if err != nil {
		err = fmt.Errorf(
			"failed to delete registration information file %q: %w",
			fileId.Path(fileType),
			err,
		)
		return
	}
	return
}

func (ri *RegInfo) Load(fileId clientstore.FileId, clientStore *clientstore.ClientStore) (err error) {
	fileType := clientstore.REG_INFO_TYPE
	riBytes, err := clientStore.ReadFile(fileId, fileType)
	if err != nil {
		err = fmt.Errorf(
			"failed to read registration information from %q: %w",
			fileId.Path(fileType),
			err,
		)
		return
	}

	parseErr := json.Unmarshal(riBytes, ri)
	if parseErr != nil {
		fmt.Fprintf(os.Stderr, "Error unmarshalling registration information JSON: %v\n", parseErr)
		os.Exit(1)
	}

	//trace("Loaded regInfo: %+v", ri)

	return
}

func (ri *RegInfo) Save(fileId clientstore.FileId, clientStore *clientstore.ClientStore) (err error) {
	//trace("Saving regInfo: %+v", ri)

	riBytes, err := json.Marshal(ri)
	if err != nil {
		err = fmt.Errorf(
			"failed to generate RegInfo JSON for %+v: %w",
			ri,
			err,
		)
		return
	}

	fileType := clientstore.REG_INFO_TYPE
	err = clientStore.WriteFile(fileId, fileType, []byte(riBytes), 0o644)
	if err != nil {
		err = fmt.Errorf(
			"failed to read registration information from %q: %w",
			fileId.Path(fileType),
			err,
		)
		return
	}

	return
}
