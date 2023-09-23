package errors

const (
	InvalidUser          = "User password must be 4 characters long and email must be 5 characters long"
	InternalServerError  = "Internal Server Error"
	InvalidCredentials   = "Invalid Credentials"
	PasswordResetEmail   = "Password reset email sent, please check your inbox"
	InvalidToken         = "Invalid Token"
	PasswordResetExpired = "Password reset token has expired"
	JsonWriteError       = "Error writing JSON"
	Unauthorized         = "Unauthorized"
)
