// Derived from cmd/public-api-demo in github.com/SUSE/connect-ng's next branch
package main

import (
	"fmt"

	"github.com/SUSE/connect-ng/pkg/connection"
)

type SccCredentials struct {
	SystemLogin string `json:"system_login"`
	Password    string `json:"password"`
	SystemToken string `json:"system_loken"`
	ShowTraces  bool   `json:"show_traces"`
}

func (SccCredentials) HasAuthentication() bool {
	return true
}

func (creds *SccCredentials) Token() (string, error) {
	if creds.ShowTraces {
		fmt.Printf("<- fetch token %s\n", creds.SystemToken)
	}
	return creds.SystemToken, nil
}

func (creds *SccCredentials) UpdateToken(token string) error {
	if creds.ShowTraces {
		fmt.Printf("-> update token %s\n", token)
	}
	creds.SystemToken = token
	return nil
}

func (creds *SccCredentials) Login() (string, string, error) {
	if creds.SystemLogin == "" || creds.Password == "" {
		return "", "", fmt.Errorf("login credentials not set")
	}

	if creds.ShowTraces {
		fmt.Printf("<- fetch login %s\n", creds.SystemLogin)
	}
	return creds.SystemLogin, creds.Password, nil
}

func (creds *SccCredentials) SetLogin(login, password string) error {
	if creds.ShowTraces {
		fmt.Printf("-> set login %s\n", login)
	}
	creds.SystemLogin = login
	creds.Password = password
	return nil
}

var _ connection.Credentials = (*SccCredentials)(nil)
