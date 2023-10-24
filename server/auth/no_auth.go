package auth

type NoAuthenticator struct{}

func (a *NoAuthenticator) GetMethod() uint8 {
	return NoAuthenticationMethod
}

func (a *NoAuthenticator) Authenticate(actual Authentication, expect Authentication) error {
	return nil
}
