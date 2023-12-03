package context_test

import (
	"fmt"
	"testing"
	"time"

	context "github.com/mcfly722/context"
)

type childNode5 struct {
	sequenceChecker sequenceChecker
}

type rootNode5 struct {
	sequenceChecker sequenceChecker
}

func (node *childNode5) Go(current context.Context) {
loop:
	for {
		select {
		case _, isOpened := <-current.Context():
			if !isOpened {
				break loop
			}
		default:
			{
			}
		}
	}
	node.sequenceChecker.NotifyWithText(7, "childNode disposing for 300ms\n")
	time.Sleep(300 * time.Millisecond)
	node.sequenceChecker.NotifyWithText(8, "childNode finished\n")
}

func (node *rootNode5) Go(current context.Context) {
loop:
	for {
		select {
		case _, isOpened := <-current.Context():
			if !isOpened {
				break loop
			}
		default:
			{
			}
		}
	}
	node.sequenceChecker.NotifyWithText(9, "rootNode finished\n")
}

func Test_FailCreateContextFromRootNode(t *testing.T) {
	sequenceChecker := newSequenceChecker()

	rootNode := &rootNode5{
		sequenceChecker: sequenceChecker,
	}

	childNode := &childNode5{
		sequenceChecker: sequenceChecker,
	}

	sequenceChecker.NotifyWithText(1, "creating new root context\n")
	rootContext := context.NewRootContext(rootNode)

	sequenceChecker.NotifyWithText(2, "creating child context\n")
	_, err := rootContext.NewContextFor(childNode)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		time.Sleep(100 * time.Millisecond)
		sequenceChecker.NotifyWithText(4, "Closing root context\n")
		rootContext.Close()

		sequenceChecker.NotifyWithText(5, "trying to create new context from closed root context...\n")

		newChildNode := &childNode5{}
		_, err := rootContext.NewContextFor(newChildNode)
		if err != nil {
			_, ok := err.(*context.ClosingIsInProcessForFreezeError)
			if ok {
				fmt.Printf("* - successfully catched error: %v\n", err)
			} else {
				panic("uncatched error")
			}
		} else {
			panic("uncatched error")
		}
	}()

	sequenceChecker.NotifyWithText(3, "Wait\n")
	rootContext.Wait()

	fmt.Printf("test finished with correct sequence = %v\n", sequenceChecker.ToString())
}
