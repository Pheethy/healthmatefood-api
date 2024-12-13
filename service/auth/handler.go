package auth

type IAuthHandler interface {
	SignToken() string
	GetExpiresAt() int
}
