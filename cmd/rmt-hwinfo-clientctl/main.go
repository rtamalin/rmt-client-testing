// Derived from cmd/public-api-demo in github.com/SUSE/connect-ng's next branch
package main

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/SUSE/connect-ng/pkg/connection"
	"github.com/SUSE/connect-ng/pkg/labels"
	"github.com/SUSE/connect-ng/pkg/registration"
	"github.com/rtamalin/rmt-client-testing/internal/clientstore"
	"github.com/rtamalin/rmt-client-testing/internal/flagtypes"
)

const (
	AppVersion = "1.0"
	Escape     = "\033"
	BoldOn     = Escape + "[1m"
	BoldOff    = Escape + "[0m"
)

func bold(format string, args ...any) {
	fmt.Printf(BoldOn+format+BoldOff+"\n", args...)
}

func langPreference() string {
	prefLang := os.Getenv("PREF_LANG")
	if prefLang == "" {
		prefLang = "en"
	}
	return prefLang
}

func loadCert(certPath string) (cert *x509.Certificate, err error) {
	crt, err := os.ReadFile(certPath)
	if err != nil {
		return
	}

	block, _ := pem.Decode(crt)
	if block == nil {
		err = fmt.Errorf("could not decode the server's certificate")
		return
	}

	return x509.ParseCertificate(block.Bytes)
}

func registerClientOld(appName, prod, version, arch, infoPath, regcode string) error {
	opts := connection.DefaultOptions(appName, AppVersion, langPreference())
	isProxy := false
	creds := &SccCredentials{}

	if url := os.Getenv("SCC_HOST"); url != "" {
		opts.URL = url
		isProxy = true
	}

	if credentialTracing := os.Getenv("TRACE_CREDENTIAL_UPDATES"); credentialTracing != "" {
		creds.ShowTraces = true
	}

	if certificatePath := os.Getenv("API_CERT"); certificatePath != "" {
		crt, certReadErr := os.ReadFile(certificatePath)
		if certReadErr != nil {
			return certReadErr
		}

		block, _ := pem.Decode(crt)
		if block == nil {
			return fmt.Errorf("Could not decode the servers certificate")
		}

		cert, parseErr := x509.ParseCertificate(block.Bytes)
		if parseErr != nil {
			return parseErr
		}

		// Set the certificate
		opts.Certificate = cert
	}

	systemInformation := registration.SystemInformation{
		"uname":    "public api demo",
		"hostname": "public-api-demo",
	}

	if infoPath != "" {
		data, readErr := os.ReadFile(infoPath)
		if readErr != nil {
			fmt.Fprintf(os.Stderr, "Error reading system information file: %v\n", readErr)
			os.Exit(1)
		}

		parseErr := json.Unmarshal(data, &systemInformation)
		if parseErr != nil {
			fmt.Fprintf(os.Stderr, "Error unmarshalling system information JSON: %v\n", parseErr)
			os.Exit(1)
		}
	}
	hostname := systemInformation["hostname"].(string)

	bold("1) Setup connection and perform an request\n")
	conn := connection.New(opts, &SccCredentials{})

	// Proxies do not implement /connect/subscriptions/info so we skip it
	if !isProxy {
		request, buildErr := conn.BuildRequest("GET", "/connect/subscriptions/info", nil)
		if buildErr != nil {
			return buildErr
		}

		connection.AddRegcodeAuth(request, regcode)

		payload, err := conn.Do(request)
		if err != nil {
			return err
		}
		fmt.Printf("!! len(payload): %d characters\n", len(payload))
		fmt.Printf("!! first 40 characters: %s\n", string(payload[0:40]))
	}

	bold("2) Registering a client to SCC with a registration code\n")
	id, regErr := registration.Register(conn, regcode, hostname, systemInformation, registration.NoExtraData)
	if regErr != nil {
		return regErr
	}
	bold("!! check https://scc.suse.com/systems/%d\n", id)

	bold("3) Activate %s/%s/%s\n", prod, version, arch)
	_, root, rootErr := registration.Activate(conn, prod, version, arch, regcode)
	if rootErr != nil {
		return rootErr
	}
	bold("++ %s activated\n", root.FriendlyName)

	bold("4) System status // Ping\n")

	extraData := registration.ExtraData{
		"instance_data": "<document>{}</document>",
	}

	systemInformation["uname"] = "public api demo - ping"

	status, statusErr := registration.Status(conn, hostname, systemInformation, extraData)
	if statusErr != nil {
		return statusErr
	}

	if status != registration.Registered {
		return errors.New("Could not finalize registration!")
	}

	bold("5) Activate recommended extensions/modules\n")
	product, prodErr := registration.FetchProductInfo(conn, prod, version, arch)
	if prodErr != nil {
		return prodErr
	}

	activator := func(ext registration.Product) (bool, error) {
		if ext.Free && ext.Recommended {
			_, act, activateErr := registration.Activate(conn, ext.Identifier, ext.Version, ext.Arch, "")
			if activateErr != nil {
				return false, activateErr
			}
			bold("++ %s activated\n", act.FriendlyName)
			return true, nil
		}
		return false, nil
	}

	if err := product.TraverseExtensions(activator); err != nil {
		return err
	}

	bold("6) Show all activations\n")
	activations, actErr := registration.FetchActivations(conn)

	if actErr != nil {
		return actErr
	}

	for i, activation := range activations {
		fmt.Printf("[%d] %s\n", i, activation.Product.Name)
	}

	bold("7) Label management\n")
	if !isProxy {
		toAssign := []labels.Label{
			{Name: "public-library-demo", Description: "Demo label created by the public-api-demo executable"},
			{Name: "to-be-removed", Description: "Demo label create by the public-api-demo-executable"},
		}

		fmt.Printf("Assigning labels..\n")
		assigned, assignErr := labels.AssignLabels(conn, toAssign)

		if assignErr != nil {
			return assignErr
		}

		fmt.Printf("Newly assigned labels:\n")
		for _, label := range assigned {
			fmt.Printf(" - %d: %s (%s)\n", label.Id, label.Name, label.Description)
		}

		index := slices.IndexFunc(assigned, func(l labels.Label) bool {
			return l.Name == "to-be-removed"
		})

		if index == -1 {
			return fmt.Errorf("Could not find to-be-removed label for this system! Something went wrong!")
		}

		fmt.Printf("Unassign %s (id: %d)..\n", assigned[index].Name, assigned[index].Id)
		_, unassignErr := labels.UnassignLabel(conn, assigned[index].Id)

		if unassignErr != nil {
			return unassignErr
		}

		fmt.Printf("Fetch updated list of labels..\n")
		updated, listErr := labels.ListLabels(conn)

		if listErr != nil {
			return listErr
		}

		fmt.Printf("Up to date list from SCC:\n")
		for _, label := range updated {
			fmt.Printf(" - %d: %s (%s)\n", label.Id, label.Name, label.Description)
		}
	} else {
		fmt.Printf("  Skipped due to running against a proxy\n")
	}

	bold("8) Deregistration of the client\n")
	if err := registration.Deregister(conn); err != nil {
		return err
	}
	bold("\n-- System deregistered\n")
	return nil
}

func readSysInfo(fileId clientstore.FileId, clientStore *clientstore.ClientStore, sysInfo *registration.SystemInformation) (err error) {
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

	parseErr := json.Unmarshal(siBytes, sysInfo)
	if parseErr != nil {
		fmt.Fprintf(os.Stderr, "Error unmarshalling system information JSON: %v\n", parseErr)
		os.Exit(1)
	}

	return
}

type RegInfo map[string]any

func readRegInfo(fileId clientstore.FileId, clientStore *clientstore.ClientStore, regInfo *RegInfo) (err error) {
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

	parseErr := json.Unmarshal(riBytes, regInfo)
	if parseErr != nil {
		fmt.Fprintf(os.Stderr, "Error unmarshalling registration information JSON: %v\n", parseErr)
		os.Exit(1)
	}

	return
}

func registerClient(id clientstore.FileId, cliOpts *CliOpts) (err error) {
	connectOpts := connection.DefaultOptions(cliOpts.appName, AppVersion, cliOpts.PrefLang)
	isProxy := false
	sccCreds := &SccCredentials{}

	if cliOpts.SccHost != "" {
		connectOpts.URL = cliOpts.SccHost
		isProxy = true
	}

	if cliOpts.TraceCreds {
		sccCreds.ShowTraces = true
	}

	if cliOpts.cert != nil {
		// Set the certificate
		connectOpts.Certificate = cliOpts.cert
	}

	sysInfo := registration.SystemInformation{
		"uname":    "public api demo",
		"hostname": "public-api-demo",
	}
	if err = readSysInfo(id, cliOpts.clientStore, &sysInfo); err != nil {
		err = fmt.Errorf(
			"registerClient failed: %w",
			err,
		)
		return
	}

	extraData := registration.ExtraData{
		"instance_data": "<document>{}</document>",
	}

	// add data profiles to extraData.dataProfiles, removing them from sysInfo
	dataProfiles := map[string]any{}
	dpNames := []string{
		"pci_data",
		"mod_data",
	}
	for _, dpName := range dpNames {
		// skip if dp_name entry not in sysInfo
		if _, ok := sysInfo[dpName]; !ok {
			continue
		}

		dataProfiles[dpName] = sysInfo[dpName]
		delete(sysInfo, dpName)
	}
	extraData["data_profiles"] = dataProfiles

	hostname := sysInfo["hostname"].(string)

	bold("Setup connection and perform registration")
	conn := connection.New(connectOpts, sccCreds)

	// Proxies do not implement /connect/subscriptions/info so we skip it
	if !isProxy {
		reqPath := "/connect/subscriptions/info"
		request, err := conn.BuildRequest("GET", reqPath, nil)
		if err != nil {
			return fmt.Errorf(
				"failed to build %s request: %w",
				reqPath,
				err,
			)
		}

		connection.AddRegcodeAuth(request, cliOpts.RegCode)

		payload, err := conn.Do(request)
		if err != nil {
			return fmt.Errorf(
				"failed to perform %s request: %w",
				reqPath,
				err,
			)
		}
		fmt.Printf("!! len(payload): %d characters\n", len(payload))
		fmt.Printf("!! first 40 characters: %s\n", string(payload[0:40]))
	}

	bold("Registering a client with %s using a registration code", connectOpts.URL)
	regId, err := registration.Register(conn, cliOpts.RegCode, hostname, sysInfo, extraData)
	if err != nil {
		err = fmt.Errorf(
			"failed to register with %q using reg code: %w",
			connectOpts.URL,
			err,
		)
		return
	}
	bold("!! check %s/systems/%d", connectOpts.URL, regId)

	bold("Activate %s/%s/%s", cliOpts.Product, cliOpts.Version, cliOpts.Arch)
	_, root, err := registration.Activate(conn, cliOpts.Product, cliOpts.Version, cliOpts.Arch, cliOpts.RegCode)
	if err != nil {
		err = fmt.Errorf(
			"failed to activate %s/%s/%s using reg code: %w",
			cliOpts.Product,
			cliOpts.Version,
			cliOpts.Arch,
			err,
		)
		return
	}
	bold("++ %s activated", root.FriendlyName)

	bold("System status // Ping")

	status, err := registration.Status(conn, hostname, sysInfo, extraData)
	if err != nil {
		err = fmt.Errorf(
			"failed to update system status for %s: %w",
			hostname,
			err,
		)
		return
	}

	if status != registration.Registered {
		return errors.New("Could not finalize registration!")
	}

	return
}

func deregisterClient(id clientstore.FileId, cliOpts *CliOpts) (err error) {
	log.Fatal("Deregistration not currently supported")
	return
}

func updateClient(id clientstore.FileId, cliOpts *CliOpts) (err error) {
	log.Fatal("Update not currently supported")
	return
}

type CliOpts struct {
	Action     CliAction
	NumClients flagtypes.Uint32
	DataStore  string
	Product    string
	Version    string
	Arch       string
	SccHost    string
	ApiCert    string
	PrefLang   string
	RegCode    string
	TraceCreds bool

	// derived values
	appName     string
	cert        *x509.Certificate
	clientStore *clientstore.ClientStore
}

var cliOpt_defaults = CliOpts{
	Action:     ACTION_REGISTER,
	NumClients: 10,
	DataStore:  "ClientDataStore",
	Product:    "SLES",
	Version:    "15.7",
	Arch:       "x86_64",
	ApiCert:    "",
	PrefLang:   "en",
}

var cliOpts CliOpts

func customTypeEnvOverride(opt flag.Value, varName, envName string) {
	// override opt value with associated env value if specified
	if value := os.Getenv(envName); value != "" {
		if err := opt.Set(value); err != nil {
			log.Fatalf(
				"Failed to set %s from %s=%q: %s",
				varName,
				envName,
				value,
				err.Error(),
			)
		}
	}
}

func stringEnvOverride(opt *string, varName, envName string) {
	if value := os.Getenv(envName); value != "" {
		*opt = value
	}
}

func boolEnvOverride(opt *bool, varName, envName string) {
	if value := os.Getenv(envName); value != "" {
		switch strings.ToLower(value) {
		case "1":
			fallthrough
		case "yes":
			fallthrough
		case "true":
			*opt = true
		default:
			*opt = false
		}
	}
}

func parseCliArgs(opts *CliOpts) {
	// initialise options from defaults
	*opts = cliOpt_defaults
	opts.appName = filepath.Base(os.Args[0])

	// override defaults with any environment settings
	customTypeEnvOverrides := []struct {
		opt     flag.Value
		varName string
		envName string
	}{
		{
			&opts.Action,
			"Action",
			"ACTION",
		},
		{
			&opts.NumClients,
			"NumClients",
			"NUM_CLIENTS",
		},
	}
	for _, o := range customTypeEnvOverrides {
		customTypeEnvOverride(o.opt, o.varName, o.envName)
	}

	stringEnvOverrides := []struct {
		opt     *string
		varName string
		envName string
	}{
		{
			&opts.DataStore,
			"DataStore",
			"DATASTORE",
		},
		{
			&opts.Product,
			"Product",
			"IDENTIFIER",
		},
		{
			&opts.Version,
			"Version",
			"VERSION",
		},
		{
			&opts.Arch,
			"Arch",
			"ARCH",
		},
		{
			&opts.SccHost,
			"SccHost",
			"SCC_HOST",
		},
		{
			&opts.ApiCert,
			"ApiCert",
			"API_CERT",
		},
		{
			&opts.PrefLang,
			"PrefLang",
			"PREF_LANG",
		},
		{
			&opts.RegCode,
			"RegCode",
			"REGCODE",
		},
	}
	for _, o := range stringEnvOverrides {
		stringEnvOverride(o.opt, o.varName, o.envName)
	}

	boolEnvOverrides := []struct {
		opt     *bool
		varName string
		envName string
	}{
		{
			&opts.TraceCreds,
			"TraceCreds",
			"TRACE_CREDENTIAL_UPDATES",
		},
	}
	for _, o := range boolEnvOverrides {
		boolEnvOverride(o.opt, o.varName, o.envName)
	}

	flag.Var(&opts.Action, "action", "Specifies the `ACTION` to perform for NUM_CLIENTS in DATASTORE.")
	flag.Var(&opts.NumClients, "clients", "The number of `NUM_CLIENTS` in DATASTORE to act upon.")
	flag.StringVar(&opts.DataStore, "datastore", opts.DataStore, "The `DATASTORE` holding the client system information JSON blobs.")
	flag.StringVar(&opts.Product, "product", opts.Product, "Register the client with this product `IDENTIFIER`.")
	flag.StringVar(&opts.Version, "version", opts.Version, "Register the client with this product `VERSION`.")
	flag.StringVar(&opts.Arch, "arch", opts.Arch, "Register the client with this product `ARCH`.")
	flag.StringVar(&opts.SccHost, "scc-host", opts.SccHost, "The `SCC_HOST` to sent requests to.")
	flag.StringVar(&opts.ApiCert, "api-cert", opts.ApiCert, "The `API_CERT` to use with specified SCC_HOST.")
	flag.StringVar(&opts.PrefLang, "lang", opts.PrefLang, "Preferred language `PREF_LANG` to use when interacting with specified SCC_HOST.")
	flag.StringVar(&opts.RegCode, "regcode", opts.RegCode, "The `REGCODE` to use when registering with specified SCC_HOST.")
	flag.BoolVar(&opts.TraceCreds, "trace", opts.TraceCreds, "Enable tracing of SCC credential operations.")

	flag.Parse()

	// sanity checks
	if opts.Action == ACTION_REGISTER && opts.RegCode == "" {
		log.Println("WARNING: No REGCODE specified for register action.")
	}

	// load the ApiCert if specified
	if opts.ApiCert != "" {
		var err error
		opts.cert, err = loadCert(opts.ApiCert)
		if err != nil {
			log.Fatalf(
				"ERROR: Failed to load specified API_CERT %q: %s",
				opts.ApiCert,
				err.Error(),
			)
		}
	}

	// setup the clientStore
	opts.clientStore = clientstore.New(opts.DataStore)
}

func performAction(id uint32, opts *CliOpts) (err error) {
	fileId := clientstore.FileId(id)
	switch opts.Action {
	case ACTION_REGISTER:
		err = registerClient(fileId, opts)
	case ACTION_UPDATE:
		err = deregisterClient(fileId, opts)
	case ACTION_DEREGISTER:
		err = updateClient(fileId, opts)
	}
	return
}

func main() {

	parseCliArgs(&cliOpts)

	for i := flagtypes.Uint32(0); i < cliOpts.NumClients; i++ {
		err := performAction(uint32(i), &cliOpts)

		if err != nil {
			log.Fatalf("Error: %s\n", err.Error())
		}
	}
}
