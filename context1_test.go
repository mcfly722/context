package context_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	context "github.com/mcfly722/context"
)

type node1 struct {
	name  string
	ready sync.Mutex
}

func (node *node1) getName() string {
	node.ready.Lock()
	defer node.ready.Unlock()
	return node.name
}

func newNode1(name string) *node1 {
	return &node1{
		name: name,
	}
}

func (node *node1) Go(current context.Context) {

	fmt.Printf("go:             %v started\n", node.getName())
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
	fmt.Printf("%v finished\n", node.getName())
}

func (parent *node1) simpleTree(context context.ChildContext, width int, height int) {

	fmt.Printf("%v configured\n", parent.getName())

	if height > 1 {
		for i := 0; i < width; i++ {
			newNode := newNode1(fmt.Sprintf("%v->%v", parent.getName(), i))
			newContext, err := context.NewContextFor(newNode)
			if err == nil {
				newNode.simpleTree(newContext, width, height-1)
			}
		}

	}

}

func Test_SimpleTree3x3(t *testing.T) {

	rootNode := newNode1("root")

	rootContext := context.NewRootContext(rootNode)

	fmt.Printf("root context created Node=%v\n", rootNode.getName())

	rootNode.simpleTree(rootContext, 3, 3)

	go func() {
		time.Sleep(10 * time.Millisecond)
		fmt.Println("Close")
		rootContext.Close()
	}()

	rootContext.Wait()
	fmt.Println("test done")

}

func Test_Ladder(t *testing.T) {

	rootNode := newNode1("root")

	rootContext := context.NewRootContext(rootNode)

	rootNode.simpleTree(rootContext, 1, 20)

	go func() {
		time.Sleep(100 * time.Millisecond)
		fmt.Println("Close")
		rootContext.Close()
	}()

	rootContext.Wait()
	fmt.Println("test done")
}
