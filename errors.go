package context

type FinishInProcessForFreezeError struct{}

func (err *FinishInProcessForFreezeError) Error() string {
	return "Finish in process. Current context state=freeze. You cannot bind new child context during closing parent context."
}

type FinishInProcessForSendError struct{}

func (err *FinishInProcessForSendError) Error() string {
	return "Could not send any control messages to closing context. Just skip this error."
}

type FinishInProcessForDisposingError struct{}

func (err *FinishInProcessForDisposingError) Error() string {
	return "Finish in process. Current context state=disposing. You cannot bind new child context during closing parent context."
}

type customPanic string

const ExitFromContextWithoutFinishPanic customPanic = "Exit from Context Without Finish() method"
