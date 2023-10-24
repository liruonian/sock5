package auth

import (
	"strings"

	"github.com/pkg/errors"
)

type UsernamePasswordAuthenticator struct{}

func (a *UsernamePasswordAuthenticator) GetMethod() uint8 {
	return UsernamePasswordAuthenticationMethod
}

func (a *UsernamePasswordAuthenticator) Authenticate(actual Authentication, expect Authentication) error {
	if strings.Compare(actual.Principle, expect.Principle) == 0 &&
		strings.Compare(actual.Credentials, expect.Credentials) == 0 {
		return nil
	}
	return errors.New("auth failed")
}
