package main

import (
	"errors"
	"fmt"

	"github.com/SUSE/connect-ng/pkg/connection"
	"github.com/SUSE/connect-ng/pkg/registration"
	"github.com/rtamalin/rmt-client-testing/internal/clientstore"
)

func deregisterClient(id clientstore.FileId, cliOpts *CliOpts) (err error) {
	connectOpts := connection.DefaultOptions(cliOpts.appName, AppVersion, cliOpts.PrefLang)
	regInfo := RegInfo{}
	sysInfo := SysInfo{}

	// load the saved system information
	err = sysInfo.Load(id, cliOpts.clientStore)
	if err != nil {
		err = errors.New(
			"deregisterClient failed to load sysInfo",
		)
		return
	}

	// retrieve the hostname from sysInfo
	hostname := sysInfo["hostname"].(string)

	// fail early if no registration info found
	if !RegInfoExists(id, cliOpts.clientStore) {
		trace("client registration missing for %q", hostname)
		err = errors.New(
			"deregisterClient client not registered",
		)
		return
	}

	// load the saved registration information
	err = regInfo.Load(id, cliOpts.clientStore)
	if err != nil {
		err = fmt.Errorf(
			"deregisterClient failed to load regInfo: %w",
			err,
		)
		return
	}

	// retrieve the client SCC creds
	sccCreds := regInfo.SccCreds

	if cliOpts.SccHost != "" {
		connectOpts.URL = cliOpts.SccHost
	}

	if cliOpts.Trace {
		sccCreds.ShowTraces = true
	}

	if cliOpts.cert != nil {
		// Set the certificate
		connectOpts.Certificate = cliOpts.cert
	}

	// we want to delete the existing registration anyway
	defer regInfo.Delete(id, cliOpts.clientStore)

	bold("Setup connection for client %q", hostname)
	conn := connection.New(connectOpts, &sccCreds)

	bold("Deregistering client %q", hostname)
	if err = registration.Deregister(conn); err != nil {
		err = fmt.Errorf(
			"deregisterClient failed to deregister a client: %w",
			err,
		)
		return
	}
	bold("Client %q deregistered", hostname)

	return
}
