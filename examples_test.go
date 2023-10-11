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
func (node *node) Go(current context.Context) {
loop:
	for {
		select {
		case _, isOpened := <-current.Context(): // this method returns context channel. If it closes, it means that we need to finish select loop
			if !isOpened {
				break loop
			}
		default: // you can use default or not, it works in both cases
			{
			}
		}
	}
	fmt.Printf("4. context '%v' closed\n", node.getName())
}

func Example() {

	rootContext := context.NewRootContext(newNode("root"))
	child1Context, _ := rootContext.NewContextFor(newNode("child1"))
	child2Context, _ := child1Context.NewContextFor(newNode("child2"))
	child2Context.NewContextFor(newNode("child3"))

	fmt.Printf("1. now waiting for 1 sec...\n")

	go func() {
		time.Sleep(1 * time.Second)
		fmt.Printf("3. one second pass\n")
		rootContext.Cancel()
	}()

	rootContext.Wait()

	fmt.Printf("5. end\n")

	// Output:
	// 1. now waiting for 1 sec...
	// 3. one second pass
	// 4. context 'child3' closed
	// 4. context 'child2' closed
	// 4. context 'child1' closed
	// 4. context 'root' closed
	// 5. end
}
