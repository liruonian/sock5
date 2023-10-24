package auth

const (
	NoAuthenticationMethod               = uint8(0)
	UsernamePasswordAuthenticationMethod = uint8(2)
)

type Authenticator interface {
	GetMethod() uint8
	Authenticate(actual Authentication, expect Authentication) error
}

type Authentication struct {
	Principle   string
	Credentials string
}
