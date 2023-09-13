package errors

const (
	InvalidUser          = "User password must be 4 characters long and email must be 5 characters long"
	InvalidPassword      = "user password must be greater than 4 characters and less than 72"
	InternalServerError  = "Internal Server Error"
	InvalidCredentials   = "Invalid Credentials"
	PasswordResetEmail   = "Password reset email sent, please check your inbox"
	InvalidToken         = "Invalid Token"
	PasswordResetExpired = "Password reset token has expired"
)
