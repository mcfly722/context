package context_test

import (
	"fmt"
	"testing"

	context "github.com/mcfly722/context"
)

type node4 struct {
	close chan context.ChildContext
}

func (node *node4) Go(current context.Context) {

	fmt.Printf("go: waiting for root context\n")
	rootContext := <-node.close

	fmt.Printf("go: root context obtained\n")

	newNode := &node4{}

	fmt.Printf("go: close context\n")
	current.Close()

	fmt.Printf("go: creating new SubContext\n")
	_, err := current.NewContextFor(newNode)
	if err != nil {
		_, ok := err.(*context.ClosingIsInProcessForDisposingError)
		if ok {
			fmt.Printf("go: successfully catched error: %v\n", err)
			rootContext.Close()
		} else {
			panic("uncatched error")
		}

	}
}

func Test_NewInstanceDuringClosing(t *testing.T) {

	rootNode := &node4{
		close: make(chan context.ChildContext),
	}

	fmt.Printf("1 - creating new context \n")
	rootContext := context.NewRootContext(rootNode)

	fmt.Printf("2 - send rootContext to node close channel\n")
	rootNode.close <- rootContext

	fmt.Printf("3 - Wait\n")
	rootContext.Wait()

	fmt.Printf("4 - test finished\n")
}
