package context

type ClosingIsInProcessForFreezeError struct{}

func (err *ClosingIsInProcessForFreezeError) Error() string {
	return "Closing is in process. Current context state=freeze. You cannot bind a new child's context during closing parent context."
}

type ClosingIsInProcessForDisposingError struct{}

func (err *ClosingIsInProcessForDisposingError) Error() string {
	return "Closing is in process. Current context state=disposing. You cannot bind a new child to context during the closing parent context."
}

type customPanic string

const ExitFromContextWithoutClosePanic customPanic = "Exit from Context Without Close() method"
