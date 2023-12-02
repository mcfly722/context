package context_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	context "github.com/mcfly722/context"
)

type node5 struct {
	name  string
	ready sync.Mutex
}

func (node *node5) getName() string {
	node.ready.Lock()
	defer node.ready.Unlock()
	return node.name
}

func newNode5(name string) *node5 {
	return &node5{
		name: name,
	}
}

func (node *node5) Go(current context.Context) {
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

func mixedLadder(path string, parents map[context.ChildContext]struct{}, width int, height int) {
	if height > 0 {
		newPath := fmt.Sprintf("%v->%v", path, height)
		newContexts := make(map[context.ChildContext]struct{})

		for i := 0; i < width; i++ {
			newInstance := newNode5(fmt.Sprintf("%v", newPath))
			fmt.Printf("%v configured\n", newInstance.getName())

			for parent := range parents {
				newContext, _ := parent.NewContextFor(newInstance)
				newContexts[newContext] = struct{}{}
			}
		}

		mixedLadder(newPath, newContexts, width, height-1)
	}
}

func Test_MixedLadder(t *testing.T) {

	rootNode := newNode5("root")

	rootContext := context.NewRootContext(rootNode)

	rootContextMap := make(map[context.ChildContext]struct{})
	rootContextMap[rootContext] = struct{}{}

	mixedLadder("root", rootContextMap, 3, 10)

	go func() {
		time.Sleep(100 * time.Millisecond)
		fmt.Println("Close")
		rootContext.Close()
	}()

	rootContext.Wait()
	fmt.Println("test done")
}
