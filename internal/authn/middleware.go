package authn

type Authenticator interface {
	Token(string) (string, error)
	Username(string) (string, error)

	Register(string) (string, error)
	Login(string) (string, error)
}
