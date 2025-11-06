package main

import (
	"errors"
	"fmt"

	"github.com/SUSE/connect-ng/pkg/connection"
	"github.com/SUSE/connect-ng/pkg/registration"
	"github.com/rtamalin/rmt-client-testing/internal/clientstore"
)

func updateClient(id clientstore.FileId, cliOpts *CliOpts) (err error) {
	connectOpts := connection.DefaultOptions(cliOpts.appName, AppVersion, cliOpts.PrefLang)
	regInfo := RegInfo{}
	sysInfo := SysInfo{}

	// load the saved system information
	err = sysInfo.Load(id, cliOpts.clientStore)
	if err != nil {
		err = fmt.Errorf(
			"updateClient failed to load sysInfo: %w",
			err,
		)
		return
	}

	// retrieve the hostname from sysInfo
	hostname := sysInfo["hostname"].(string)

	// fail early if no registration info found
	if !RegInfoExists(id, cliOpts.clientStore) {
		trace("client registration missing for %q", hostname)
		err = errors.New(
			"updateClient client not registered",
		)
		return
	}

	// load the saved registration information
	err = regInfo.Load(id, cliOpts.clientStore)
	if err != nil {
		err = fmt.Errorf(
			"updateClient failed to load regInfo: %w",
			err,
		)
		return
	}

	// retrieve the client SCC creds
	sccCreds := regInfo.SccCreds

	// generate the client's extraData
	extraData := prepareExtraData(sysInfo, cliOpts)

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

	trace("Setup connection for client %q", hostname)
	conn := connection.New(connectOpts, &sccCreds)

	trace("Sending keepalive heartbeat for client %q", hostname)
	status, err := registration.Status(conn, hostname, sysInfo, extraData)
	if err != nil {
		err = fmt.Errorf(
			"updateClient failed to update system status for %q: %w",
			hostname,
			err,
		)
		return
	}

	if status != registration.Registered {
		trace("heartbeat failed as client %q not registered", hostname)
		// delete the existing regInfo
		_ = regInfo.Delete(id, cliOpts.clientStore)

		err = errors.New("failed to send keepalive heartbeat")
		err = fmt.Errorf(
			"updateClient failed to update system status for %q: %w",
			hostname,
			err,
		)
		return
	}

	regInfo.SccCreds = sccCreds
	err = regInfo.Save(id, cliOpts.clientStore)
	if err != nil {
		err = fmt.Errorf(
			"updatedClient failed to save updated registration info: %w",
			err,
		)
	}

	bold("Client %q keepalive heartbeat updated", hostname)

	return
}
