package main

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

// enable to show trace messages
var traceEnabled = false

func configTracing(enableTracing bool) {
	traceEnabled = enableTracing
}

func bold(format string, args ...any) {
	fmt.Printf(BoldOn+format+BoldOff+"\n", args...)
}

func trace(format string, args ...any) {
	if traceEnabled {
		fmt.Printf(TracePrefix+format+"\n", args...)
	}
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

func loadInstData(instDataPath string) (instData string, err error) {
	instDataBytes, err := os.ReadFile(instDataPath)
	if err != nil {
		return
	}

	instData = string(instDataBytes)
	return
}
