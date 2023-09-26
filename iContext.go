package context

// Context ...
type Context interface {
	// create new child context for instance what implements Instance interface
	NewContextFor(instance Instance) error

	// channel what closes when all childs are closed and you can exit from your current context. Dispose would be called.
	CurrentContextIsOpened() chan struct{}

	// finish all sub*childs, childs and current context
	GracefulFinish()

	// set debugger
	SetDebugger(debugger func(message DebuggerMessage))
}

// RootContext ...
type RootContext interface {
	// Create new child context
	NewContextFor(instance ContextedInstance) (Context, error)

	// Graceful finish all sub*childs, then childs and after current root context
	GracefulFinish()

	// wait till this root context would be closed
	Wait()

	// set debugger
	SetDebugger(debugger func(message DebuggerMessage))
}

// Instance ...
type Instance interface {
	// Goroutine with select loop to serve all incomming messages from childs and CurrentContextIsOpened channel
	Go(current Context)

	// after
	Dispose()
}

type DebuggerMessage string

const (
	NewContext         = "NewContext"
	GracefulFinishCall = "GracefulFinish"
	Freeze             = "Freeze"
	FreezeEnd          = "FreezeEnd"
	Dispose            = "Dispose"
	DisposeEnd         = "DisposeEnd"
)

// ParentContextAlreadyInClosingStateError ...
type GracefulShutdownInProcessError struct{}

func (err *GracefulShutdownInProcessError) Error() string {
	return "graceful shutdown in process. You cannot bind new childs during closing parent context. Just exit."
}

// GracefulFinishFromDisposeStateError ...
type GracefulFinishFromDisposeStateError struct{}

func (err *GracefulFinishFromDisposeStateError) Error() string {
	return "GracefulFinish() cloud not be called from Dispose state."
}

// ParentContextTriesToExitWithoutGracefulFinishError ...
type ParentContextTriesToExitWithoutGracefulFinishError struct{}

func (err *ParentContextTriesToExitWithoutGracefulFinishError) Error() string {
	return "Parent context tries to Exit without GracefulFinish()"
}
