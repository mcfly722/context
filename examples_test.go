package context_test

import (
	"fmt"
	"time"

	context "github.com/mcfly722/context"
)

type node struct {
	name string
}

func newNode(name string) *node {
	return &node{name: name}
}

func (node *node) getName() string {
	return node.name
}

// this method node should implement as Goroutine loop
func (node *node) Go(current context.Context[any]) {
loop:
	for {
		select {
		case _, isOpened := <-current.Controller(): // this method returns context channel. If it finished, it means that we need exit from select loop and function
			if !isOpened {
				break loop
			}
		default: // you can use default or not, it works in both cases
			{
			}
		}
	}
	fmt.Printf("4. context '%v' finished\n", node.getName())
}

func Example() {

	rootContext := context.NewRootContext[any](newNode("root"))
	child1Context, _ := rootContext.NewContextFor(newNode("child1"))
	child2Context, _ := child1Context.NewContextFor(newNode("child2"))
	child2Context.NewContextFor(newNode("child3"))

	fmt.Printf("1. now waiting for 1 sec...\n")

	go func() {
		time.Sleep(1 * time.Second)
		fmt.Printf("3. one second pass\n")
		rootContext.Finish()
	}()

	rootContext.Wait()

	fmt.Printf("5. end\n")

	// Output:
	// 1. now waiting for 1 sec...
	// 3. one second pass
	// 4. context 'child3' finished
	// 4. context 'child2' finished
	// 4. context 'child1' finished
	// 4. context 'root' finished
	// 5. end
}
