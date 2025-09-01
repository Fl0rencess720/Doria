package response

const (
	UserExistError = iota + 5
	CodeError

	UserNotExistError
	PasswordError

	OtherError = 1000
	Success    = 2000
)
