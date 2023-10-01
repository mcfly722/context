package context

type treeMessage struct {
	context *context
	result  chan error
}

func sendContextTo(channel chan *treeMessage, context *context) error {

	msg := &treeMessage{
		context: context,
		result:  make(chan error),
	}

	channel <- msg
	return <-msg.result
}

func (treeMessage *treeMessage) answer(getAnswer func() error) {
	err := getAnswer()
	treeMessage.result <- err
}
