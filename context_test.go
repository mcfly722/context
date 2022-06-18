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
			current.Log(102, "close signal")
			break loop
		case <-current.OnDone():
			break loop
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
	fmt.Println("waiting for closing")
	ctx.Wait()

}

func Test_ImmediateExitFromFirstChild(t *testing.T) {
	fmt.Println("correct closing?")

	root := context.NewRootContext(context.NewConsoleLogDebugger())

	node0 := newNode()
	node1 := newNode()

	ctx0 := root.NewContextFor(node0, "0", "node")
	ctx0.NewContextFor(node1, "1", "node")

	go func() {
		time.Sleep(3 * time.Second)
		fmt.Println("correct closing!")
		node0.close <- true
	}()

	ctx0.Wait()
}

func Test_ImmediateExitFromRoot(t *testing.T) {
	fmt.Println("correct closing?")

	root := context.NewRootContext(context.NewConsoleLogDebugger())

	node1 := newNode()

	root.NewContextFor(node1, "1", "node")

	go func() {
		root.Log(0, "startedGoRoutine")
		time.Sleep(3 * time.Second)
		root.Log(0, "correct closing!")
		root.Terminate()
	}()

	root.Wait()
	root.Log(0, "Test_ImmediateExitFromRoot finished")

}

func Test_EmptyDebugger(t *testing.T) {
	root := context.NewRootContext(context.NewEmptyDebugger())

	go func() {
		fmt.Println("startedGoRoutine")
		time.Sleep(3 * time.Second)
		fmt.Println("correct closing!")
		root.Terminate()
	}()

	root.Wait()
}

func Test_AddToRootAfterTermination(t *testing.T) {

	root := context.NewRootContext(context.NewConsoleLogDebugger())

	go func() {
		root.Terminate()
		root.Terminate()
		root.Terminate()
	}()

	node := newNode()
	ctx := root.NewContextFor(node, "0", "node")
	node.buildContextTree(ctx, 2, 5)

	go func() {
		fmt.Println("closing!")
		root.Terminate()
	}()

	root.Wait()
}
