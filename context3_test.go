package context_test

import (
	"testing"

	context "github.com/mcfly722/context"
)

type node4 struct {
	close           chan context.ChildContext
	sequenceChecker sequenceChecker
}

func (node *node4) Go(current context.Context) {

	node.sequenceChecker.NotifyWithText(3, "3 - waiting for root context\n")
	rootContext := <-node.close

	node.sequenceChecker.NotifyWithText(4, "4 - root context obtained\n")
	newNode := &node4{}

	node.sequenceChecker.NotifyWithText(5, "5 - close context\n")
	current.Close()

	node.sequenceChecker.NotifyWithText(6, "6 - creating new SubContext\n")
	_, err := current.NewContextFor(newNode)
	if err != nil {
		_, ok := err.(*context.ClosingIsInProcessForDisposingError)
		if ok {
			node.sequenceChecker.NotifyWithText(7, "7 - successfully catched error: %v\n", err)
			rootContext.Close()
		} else {
			panic("uncatched error")
		}

	}
}

func Test_NewInstanceDuringClosing(t *testing.T) {
	sequenceChecker := newSequenceChecker()

	rootNode := &node4{
		close:           make(chan context.ChildContext),
		sequenceChecker: sequenceChecker,
	}

	sequenceChecker.NotifyWithText(1, "1 - creating new context \n")
	rootContext := context.NewRootContext(rootNode)

	sequenceChecker.NotifyWithText(2, "2 - send rootContext to node close channel\n")
	rootNode.close <- rootContext

	rootContext.Wait()

	sequenceChecker.NotifyWithText(8, "test finished with correct sequence = %v\n", sequenceChecker.ToString())
}
