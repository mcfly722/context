package context

// CancelInProcessError ...
type CancelInProcessError struct{}

func (err *CancelInProcessError) Error() string {
	return "Cancel in process. You cannot bind new childs during closing parent context."
}

type customPanic string

const (
	ExitFromContextWithoutCancelPanic customPanic = "Exit from Context Without Cancel() method"
)
