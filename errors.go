package context

// ParentContextAlreadyInClosingStateError ...
type GracefulShutdownInProcessError struct{}

func (err *GracefulShutdownInProcessError) Error() string {
	return "graceful shutdown in process. You cannot bind new childs during closing parent context. Just exit."
}

// ParentContextTriesToExitWithoutGracefulFinishError ...
type ParentContextTriesToExitWithoutGracefulFinishError struct{}

func (err *ParentContextTriesToExitWithoutGracefulFinishError) Error() string {
	return "Parent context tries to Exit without GracefulFinish()"
}

// NewContextInstanceFromDisposingStateError ...
type NewContextInstanceFromDisposingStateError struct{}

func (err *NewContextInstanceFromDisposingStateError) Error() string {
	return "Could not create new context for disposing context"
}

// NewContextInstanceFromFreezedStateError ...
type NewContextInstanceFromFreezedStateError struct{}

func (err *NewContextInstanceFromFreezedStateError) Error() string {
	return "Could not create new context from freezed context"
}

// GracefulFinishFromDisposeStateError ...
type GracefulFinishFromDisposeStateError struct{}

func (err *GracefulFinishFromDisposeStateError) Error() string {
	return "GracefulFinish() cloud not be called from dispose state"
}
