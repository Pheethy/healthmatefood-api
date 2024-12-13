package constants

const (
	ERROR_USER_NOT_FOUND           = "user not found"
	ERROR_USERNAME_WAS_DUPLICATED  = "username was duplicated"
	ERROR_EMAIL_WAS_DUPLICATED     = "email was duplicated"
	ERROR_EMAIL_PATTERN_IS_INVALID = "email pattern is invalid"
	ERROR_PASSWORD_IS_INVALID      = "password is invalid"
	ERROR_OAUTH_NOT_FOUND          = "oauth not found"
)

const (
	POSTGRES_ERROR_USERNAME_WAS_DUPLICATED = "pq: duplicate key value violates unique constraint \"users_username_unique\""
	POSTGRES_ERROR_EMAIL_WAS_DUPLICATED    = "pq: duplicate key value violates unique constraint \"users_email_unique\""
)

type ErrorResponse struct {
	Message string `json:"message" example:"Invalid email format"`
	Code    int    `json:"code" example:"400"`
}
