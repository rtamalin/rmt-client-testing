package main

import (
	"errors"
	"fmt"

	"github.com/SUSE/connect-ng/pkg/connection"
	"github.com/SUSE/connect-ng/pkg/registration"
	"github.com/rtamalin/rmt-client-testing/internal/clientstore"
)

func registerClient(id clientstore.FileId, cliOpts *CliOpts) (err error) {
	connectOpts := connection.DefaultOptions(cliOpts.appName, AppVersion, cliOpts.PrefLang)
	isProxy := false
	sccCreds := SccCredentials{}

	// load the client's system information
	sysInfo := SysInfo{
		"uname":    "public api demo",
		"hostname": "public-api-demo",
	}
	if err = sysInfo.Load(id, cliOpts.clientStore); err != nil {
		err = fmt.Errorf(
			"registerClient failed to load system information: %w",
			err,
		)
		return
	}

	// generate the client's extraData
	extraData := extraDataWithDataProfiles(sysInfo, cliOpts)

	if cliOpts.SccHost != "" {
		connectOpts.URL = cliOpts.SccHost
		isProxy = true
	}

	if cliOpts.Trace {
		sccCreds.ShowTraces = true
	}

	if cliOpts.cert != nil {
		// Set the certificate
		connectOpts.Certificate = cliOpts.cert
	}

	// retrieve the hostname from sysInfo
	hostname := sysInfo["hostname"].(string)

	// fail if attempting to register a client that already exists
	if RegInfoExists(id, cliOpts.clientStore) {
		trace("client registration already exists for %q", hostname)
		err = errors.New(
			"registerClient client already registered",
		)
		return
	}

	bold("Setup connection for client %q", hostname)
	conn := connection.New(connectOpts, &sccCreds)

	// Proxies do not implement /connect/subscriptions/info so we skip it
	if !isProxy {
		reqPath := "/connect/subscriptions/info"
		request, err := conn.BuildRequest("GET", reqPath, nil)
		if err != nil {
			return fmt.Errorf(
				"registerClient failed to build %s request: %w",
				reqPath,
				err,
			)
		}

		connection.AddRegcodeAuth(request, cliOpts.RegCode)

		payload, err := conn.Do(request)
		if err != nil {
			return fmt.Errorf(
				"registerClient failed to perform %s request: %w",
				reqPath,
				err,
			)
		}
		trace("len(payload): %d characters", len(payload))
		trace("first 40 characters: %s", string(payload[0:40]))
	}

	bold("Registering client %q against %q using a registration code", hostname, connectOpts.URL)
	regId, err := registration.Register(conn, cliOpts.RegCode, hostname, sysInfo, extraData)
	if err != nil {
		err = fmt.Errorf(
			"registerClient failed to register with %q using reg code: %w",
			connectOpts.URL,
			err,
		)
		return
	}
	trace("check %s/systems/%d", connectOpts.URL, regId)

	bold("Activating %s/%s/%s for client %q", cliOpts.Product, cliOpts.Version, cliOpts.Arch, hostname)
	_, root, err := registration.Activate(conn, cliOpts.Product, cliOpts.Version, cliOpts.Arch, cliOpts.RegCode)
	if err != nil {
		err = fmt.Errorf(
			"registerClient failed to activate %s/%s/%s using reg code: %w",
			cliOpts.Product,
			cliOpts.Version,
			cliOpts.Arch,
			err,
		)
		// deregister the client if the activation fails
		_ = registration.Deregister(conn)
		return
	}
	trace("%s activated for client %q", root.FriendlyName, hostname)

	// record registration info
	regInfo := RegInfo{
		SccCreds: sccCreds,
	}

	err = regInfo.Save(id, cliOpts.clientStore)
	if err != nil {
		err = fmt.Errorf(
			"registerClient failed to save registration info: %w",
			err,
		)
	}

	return
}
