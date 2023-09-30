package context

// CancelInProcessError ...
type CancelInProcessError struct{}

func (err *CancelInProcessError) Error() string {
	return "Cancel in process. You cannot bind new childs during closing parent context."
}

type customPanic string

const (
	CancelFromDisposeStatePanic       customPanic = "Cancel() cloud not be called from dispose state"
	ExitFromContextWithoutCancelPanic customPanic = "Exit from Context Without Cancel() method"
)
