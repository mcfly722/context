package context_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/mcfly722/goPackages/context"
)

type node struct {
	close chan bool
}

type debugger struct{}

func newNode() *node {
	return &node{
		close: make(chan bool),
	}
}

func (node *node) Go(current context.Context) {

loop:
	for {
		select {
		case <-node.close:
			//fmt.Printf("Cancel()")
			current.Cancel()
			break
		case _, opened := <-current.Opened():
			if !opened {
				break loop
			}
		}

	}
}

func (node *node) Dispose(current context.Context) {}

func (parent *node) buildContextTree(parentCtx context.Context, width int, depth int) context.Context {

	if depth > 0 {
		for i := 0; i < width; i++ {
			newChildNode := newNode()
			newChildContext := parentCtx.NewContextFor(newChildNode, fmt.Sprintf("%v", i), "node")
			newChildNode.buildContextTree(newChildContext, width, depth-1)
		}
	}

	return nil
}

func Test_SimpleTree(t *testing.T) {

	node := newNode()

	root := context.NewRootContext(context.NewConsoleLogDebugger())
	ctx := root.NewContextFor(node, "0", "node")

	node.buildContextTree(ctx, 2, 5)

	time.Sleep(1 * time.Second)
	fmt.Println("send cancel")
	node.close <- true

	go func() {
		time.Sleep(1 * time.Second)
		root.Cancel()
	}()

	root.Wait()

}
