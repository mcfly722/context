package context_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/mcfly722/goPackages/context"
)

type node struct {
	name  string
	close chan bool
}

func newNode(name string) *node {
	//fmt.Println(fmt.Sprintf("[%v] created", name))

	return &node{
		name:  name,
		close: make(chan bool),
	}
}

func (node *node) GoContextBody(current context.Context) {

	fmt.Println(fmt.Sprintf("[%v] started", node.name))

loop:
	for {
		select {
		case <-node.close:
			break loop
		case <-current.OnDone():
			break loop
		}

	}

	//fmt.Println(fmt.Sprintf("[%v] finished", node.name))
}

func (node *node) Dispose() {
	fmt.Println(fmt.Sprintf("[%v] disposed", node.name))
}

func (parent *node) buildContextTree(parentCtx context.Context, width int, depth int) context.Context {

	if depth > 0 {
		for i := 0; i < width; i++ {
			newChildNode := newNode(fmt.Sprintf("%v->%v", parent.name, i))

			newChildContext := parentCtx.NewContextFor(newChildNode)

			newChildNode.buildContextTree(newChildContext, width, depth-1)
		}
	}

	return nil
}

func Test_SimpleTree(t *testing.T) {

	node := newNode("0")

	ctx := context.NewContextFor(node)

	node.buildContextTree(ctx, 2, 5)

	time.Sleep(1 * time.Second)
	fmt.Println("send cancel")
	node.close <- true
	fmt.Println("waiting for closing")
	ctx.Wait()

}
