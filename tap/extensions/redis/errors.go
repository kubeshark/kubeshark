package main

//RedisError basic redis error
type RedisError struct {
	Message string
}

func newRedisError(message string) *RedisError {
	return &RedisError{Message: message}
}

func (e *RedisError) Error() string {
	return e.Message
}

//RedirectError cluster operation redirect error
type RedirectError struct {
	Message string
}

func newRedirectError(message string) *RedirectError {
	return &RedirectError{Message: message}
}

func (e *RedirectError) Error() string {
	return e.Message
}

//ClusterMaxAttemptsError cluster operation exceed max attempts errror
type ClusterMaxAttemptsError struct {
	Message string
}

func newClusterMaxAttemptsError(message string) *ClusterMaxAttemptsError {
	return &ClusterMaxAttemptsError{Message: message}
}

func (e *ClusterMaxAttemptsError) Error() string {
	return e.Message
}

//NoReachableClusterNodeError have no reachable cluster node error
type NoReachableClusterNodeError struct {
	Message string
}

func newNoReachableClusterNodeError(message string) *NoReachableClusterNodeError {
	return &NoReachableClusterNodeError{Message: message}
}

func (e *NoReachableClusterNodeError) Error() string {
	return e.Message
}

//MovedDataError cluster move data error
type MovedDataError struct {
	Message string
	Host    string
	Port    int
	Slot    int
}

func newMovedDataError(message string, host string, port int, slot int) *MovedDataError {
	return &MovedDataError{Message: message, Host: host, Port: port, Slot: slot}
}

func (e *MovedDataError) Error() string {
	return e.Message
}

//AskDataError ask data error
type AskDataError struct {
	Message string
	Host    string
	Port    int
	Slot    int
}

func newAskDataError(message string, host string, port int, slot int) *AskDataError {
	return &AskDataError{Message: message, Host: host, Port: port, Slot: slot}
}

func (e *AskDataError) Error() string {
	return e.Message
}

//ClusterError cluster basic error
type ClusterError struct {
	Message string
}

func newClusterError(message string) *ClusterError {
	return &ClusterError{Message: message}
}

func (e *ClusterError) Error() string {
	return e.Message
}

//BusyError operation is busy error
type BusyError struct {
	Message string
}

func newBusyError(message string) *BusyError {
	return &BusyError{Message: message}
}

func (e *BusyError) Error() string {
	return e.Message
}

//NoScriptError has no script error
type NoScriptError struct {
	Message string
}

func newNoScriptError(message string) *NoScriptError {
	return &NoScriptError{Message: message}
}

func (e *NoScriptError) Error() string {
	return e.Message
}

//DataError data error
type DataError struct {
	Message string
}

func newDataError(message string) *DataError {
	return &DataError{Message: message}
}

func (e *DataError) Error() string {
	return e.Message
}

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

//ClusterOperationError cluster operation error
type ClusterOperationError struct {
	Message string
}

func newClusterOperationError(message string) *ClusterOperationError {
	return &ClusterOperationError{Message: message}
}

func (e *ClusterOperationError) Error() string {
	return e.Message
}
