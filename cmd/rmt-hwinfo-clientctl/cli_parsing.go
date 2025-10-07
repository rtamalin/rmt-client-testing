package main

import (
	"crypto/x509"
	"flag"
	"log"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/rtamalin/rmt-client-testing/internal/clientstore"
)

type CliOpts struct {
	Action       CliAction
	NumClients   int64
	NumJobs      int64
	DataStore    string
	Product      string
	Version      string
	Arch         string
	SccHost      string
	ApiCert      string
	PrefLang     string
	RegCode      string
	InstDataPath string
	Trace        bool

	// derived values
	appName     string
	cert        *x509.Certificate
	clientStore *clientstore.ClientStore
	instData    string
}

var cliOpt_defaults = CliOpts{
	Action:     ACTION_REGISTER,
	NumClients: 10,
	NumJobs:    10,
	DataStore:  "ClientDataStore",
	Product:    "SLES",
	Version:    "15.7",
	Arch:       "x86_64",
	PrefLang:   langPreference(),
	instData:   "<document>{}</document>",
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

func int64EnvOverride(opt *int64, varName, envName string) {
	if value := os.Getenv(envName); value != "" {
		val, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			log.Fatalf(
				"Failed to set %s from %s=%q: %s",
				varName,
				envName,
				value,
				err.Error(),
			)
		}
		*opt = val
	}
}

func stringEnvOverride(opt *string, _, envName string) {
	if value := os.Getenv(envName); value != "" {
		*opt = value
	}
}

func boolEnvOverride(opt *bool, _, envName string) {
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
	}
	for _, o := range customTypeEnvOverrides {
		customTypeEnvOverride(o.opt, o.varName, o.envName)
	}

	int64EnvOverrides := []struct {
		opt     *int64
		varName string
		envName string
	}{
		{
			&opts.NumClients,
			"NumClients",
			"NUM_CLIENTS",
		},
		{
			&opts.NumJobs,
			"NumJobs",
			"NUM_JOBS",
		},
	}
	for _, o := range int64EnvOverrides {
		int64EnvOverride(o.opt, o.varName, o.envName)
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
		{
			&opts.InstDataPath,
			"InstanceData",
			"INST_DATA",
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
			&opts.Trace,
			"Trace",
			"TRACE_UPDATES",
		},
	}
	for _, o := range boolEnvOverrides {
		boolEnvOverride(o.opt, o.varName, o.envName)
	}

	flag.Var(&opts.Action, "action", "Specifies the `ACTION` to perform for NUM_CLIENTS in DATASTORE.")
	flag.Int64Var(&opts.NumClients, "clients", opts.NumClients, "`NUM_CLIENTS` specifies the number client entries in DATASTORE to act upon.")
	flag.Int64Var(&opts.NumJobs, "jobs", opts.NumJobs, "`NUM_JOBS` to run in parallel.")
	flag.StringVar(&opts.DataStore, "datastore", opts.DataStore, "The `DATASTORE` holding the client system information JSON blobs.")
	flag.StringVar(&opts.Product, "product", opts.Product, "Register the client with this product `IDENTIFIER`.")
	flag.StringVar(&opts.Version, "version", opts.Version, "Register the client with this product `VERSION`.")
	flag.StringVar(&opts.Arch, "arch", opts.Arch, "Register the client with this product `ARCH`.")
	flag.StringVar(&opts.SccHost, "scc-host", opts.SccHost, "The `SCC_HOST` to sent requests to.")
	flag.StringVar(&opts.ApiCert, "api-cert", opts.ApiCert, "The `API_CERT` to use with specified SCC_HOST.")
	flag.StringVar(&opts.PrefLang, "lang", opts.PrefLang, "Preferred language `PREF_LANG` to use when interacting with specified SCC_HOST.")
	flag.StringVar(&opts.RegCode, "regcode", opts.RegCode, "The `REGCODE` to use when registering with specified SCC_HOST.")
	flag.StringVar(&opts.InstDataPath, "instdata", opts.InstDataPath, "The `INST_DATA` to use when registering with specified SCC_HOST.")
	flag.BoolVar(&opts.Trace, "trace", opts.Trace, "Enable tracing of operations.")

	flag.Parse()

	//
	// sanity checks
	//

	// fail if the number of clients is invalid
	if (opts.NumClients >= math.MaxUint32) || (opts.NumClients < 0) {
		log.Fatal(
			"ERROR: The number of clients must be a positive value between 0 and MaxUint32\n",
		)
	}

	// fail if the number of parallel jobs is invalid
	if (opts.NumJobs >= math.MaxUint32) || (opts.NumJobs < 0) {
		log.Fatal(
			"ERROR: The number of parallel must be a positive value between 0 and MaxUint32\n",
		)
	}

	// warn if trying to register without specifying REGCODE or INST_DATA
	if opts.Action == ACTION_REGISTER &&
		(opts.RegCode == "") && (opts.InstDataPath == "") {
		log.Println("WARNING: No REGCODE or INST_DATA specified for register action.")
	}

	// configure tracing
	configTracing(opts.Trace)

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

	// load the instance data is specified
	if opts.InstDataPath != "" {
		var err error
		opts.instData, err = loadInstData(opts.InstDataPath)
		if err != nil {
			log.Fatalf(
				"ERROR: Failed to load specified INST_DATA %q: %s",
				opts.InstDataPath,
				err.Error(),
			)
		}
	}

	// setup the clientStore
	opts.clientStore = clientstore.New(opts.DataStore)
}
