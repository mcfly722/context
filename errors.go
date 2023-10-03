package context

type CancelInProcessForFreezeError struct{}

func (err *CancelInProcessForFreezeError) Error() string {
	return "Cancel in process. Current context state=freeze. You cannot bind new child context during closing parent context."
}

type CancelInProcessForDisposingError struct{}

func (err *CancelInProcessForDisposingError) Error() string {
	return "Cancel in process. Current context state=disposing. You cannot bind new child context during closing parent context."
}

type customPanic string

const ExitFromContextWithoutCancelPanic customPanic = "Exit from Context Without Cancel() method"
