package context_test

import (
	"fmt"
	"testing"

	context "github.com/mcfly722/context"
)

type node4 struct {
	cancel chan context.ContextNode
}

func (node *node4) Go(current context.Context) {

	fmt.Printf("go: waiting for root context\n")
	rootContext := <-node.cancel

	fmt.Printf("go: root context obtained\n")

	newNode := &node4{}

	fmt.Printf("go: cancel context\n")
	current.Cancel()

	fmt.Printf("go: creating new SubContext\n")
	_, err := current.NewContextFor(newNode)
	if err != nil {
		_, ok := err.(*context.CancelInProcessForDisposingError)
		if ok {
			fmt.Printf("go: successfully catched error: %v\n", err)
			rootContext.Cancel()
		} else {
			panic("uncatched error")
		}

	}
}

func Test_NewInstanceDuringCancel(t *testing.T) {

	rootNode := &node4{
		cancel: make(chan context.ContextNode),
	}

	fmt.Printf("1 - creating new context \n")
	rootContext := context.NewRootContext(rootNode)

	fmt.Printf("2 - send rootContext to node cancel channel\n")
	rootNode.cancel <- rootContext

	fmt.Printf("3 - Wait\n")
	rootContext.Wait()

	fmt.Printf("4 - test finished\n")
}
