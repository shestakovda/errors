package errx

// Standard HTTP errors
var (
	// 4хх
	ErrBadRequest    = New("400 Bad Request")
	ErrUnauthorized  = New("401 Unauthorized")
	ErrForbidden     = New("403 Forbidden")
	ErrNotFound      = New("404 Not Found")
	ErrNotAllowed    = New("405 Method Not Allowed")
	ErrNotAcceptable = New("406 Not Acceptable")
	ErrUnprocessable = New("422 Unprocessable Entity")

	// 5хх
	ErrInternal       = New("500 Internal Server Error")
	ErrNotImplemented = New("501 Not Implemented")
	ErrUnavailable    = New("503 Service Unavailable")
)
