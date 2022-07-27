package redis

//ConnectError redis connection error,such as io timeout
type ConnectError struct {
	Message string
}

func newConnectError(message string) *ConnectError {
	return &ConnectError{Message: message}
}

func (e *ConnectError) Error() string {
	return e.Message
}
