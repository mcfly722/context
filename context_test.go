package context_test

import (
	"fmt"
	"testing"
	"time"

	context "github.com/mcfly722/context"
)

type node1 struct {
	i    int
	path string
}

func (node *node1) name() string {
	return (fmt.Sprintf("%v->%v", node.path, node.i))
}

func newNode1(path string, i int) *node1 {

	return &node1{
		path: path,
		i:    i,
	}
}

func (node *node1) Go(current context.Context) {
loop:
	for {
		select {
		case _, opened := <-current.IsOpen():
			if !opened {
				break loop
			}
		default:
			{
			}
		}
	}
	fmt.Printf("%v finished\n", node.name())
}

func (parent *node1) simpleTree(context context.ContextNode, width int, height int) {
	if height > 1 {

		for i := 0; i < width; i++ {

			newNode := newNode1(fmt.Sprintf("%v->%v", parent.path, i), i)
			newContext, err := context.NewContextFor(newNode)
			if err == nil {
				fmt.Printf("%v started\n", newNode.name())
				newNode.simpleTree(newContext, width, height-1)
			}

		}

	}

}

func Test_SimpleTree3x3(t *testing.T) {

	rootNode := newNode1("", 0)

	rootContext, err := context.NewRootContext(rootNode)
	if err != nil {
		t.Fatal(err)
	}

	rootNode.simpleTree(rootContext, 3, 3)

	go func() {
		time.Sleep(10 * time.Millisecond)
		fmt.Println("Cancel")
		rootContext.Cancel()
	}()

	rootContext.Wait()
	fmt.Println("Finished")

}

func Test_Ladder(t *testing.T) {

	rootNode := newNode1("", 0)

	rootContext, err := context.NewRootContext(rootNode)
	if err != nil {
		t.Fatal(err)
	}

	rootNode.simpleTree(rootContext, 1, 20)

	go func() {
		time.Sleep(100 * time.Millisecond)
		fmt.Println("Cancel")
		rootContext.Cancel()
	}()

	rootContext.Wait()
	fmt.Println("Finished")
}
