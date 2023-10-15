package context_test

import (
	"fmt"
	"testing"

	context "github.com/mcfly722/context"
)

type node4 struct {
	Close chan context.ChildContext[any]
}

func (node *node4) Go(current context.Context[any]) {

	fmt.Printf("go: waiting for root context\n")
	rootContext := <-node.Close

	fmt.Printf("go: root context obtained\n")

	newNode := &node4{}

	fmt.Printf("go: Close context\n")
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

func Test_NewInstanceDuringClose(t *testing.T) {

	rootNode := &node4{
		Close: make(chan context.ChildContext[any]),
	}

	fmt.Printf("1 - creating new context \n")
	rootContext := context.NewRootContext[any](rootNode)

	fmt.Printf("2 - send rootContext to node Close channel\n")
	rootNode.Close <- rootContext

	fmt.Printf("3 - Wait\n")
	rootContext.Wait()

	fmt.Printf("4 - test closed\n")
}
