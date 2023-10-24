package server

const (
	socks5Version = uint8(5)

	authVersion = uint8(1)
	authSuccess = uint8(0)
	authFailure = uint8(1)

	ipv4Address = uint8(1)
	fqdnAddress = uint8(3)
	ipv6Address = uint8(4)

	ConnectCommand   = uint8(1)
	BindCommand      = uint8(2)
	AssociateCommand = uint8(3)
)

const (
	succeeded uint8 = iota
	generalSocksServerFailure
	connectionNotAllowedByRuleset
	networkUnreachable
	hostUnreachable
	connectionRefused
	ttlExpired
	commandNotSupported
	addressTypeNotSupported
)
