package context_test

import (
	"fmt"
	"testing"
	"time"

	context "github.com/mcfly722/context"
)

type childNode5 struct{}
type rootNode5 struct{}

func (node *childNode5) Go(current context.Context[any]) {
loop:
	for {
		select {
		case _, isOpened := <-current.Controller():
			if !isOpened {
				break loop
			}
		default:
			{
			}
		}
	}
	fmt.Printf("go:     childNode disposing for 300ms\n")
	time.Sleep(300 * time.Millisecond)
	fmt.Printf("go:     childNode finished\n")
}

func (node *rootNode5) Go(current context.Context[any]) {
loop:
	for {
		select {
		case _, isOpened := <-current.Controller():
			if !isOpened {
				break loop
			}
		default:
			{
			}
		}
	}
	fmt.Printf("go:     rootNode finished\n")
}

func Test_FailCreateContextFromRootNode(t *testing.T) {

	rootNode := &rootNode5{}
	childNode := &childNode5{}

	fmt.Printf("1 - creating new root context\n")
	rootContext := context.NewRootContext[any](rootNode)

	fmt.Printf("2 - creating child context\n")
	_, err := rootContext.NewContextFor(childNode)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		time.Sleep(100 * time.Millisecond)
		fmt.Printf("4 - Finishing root context\n")
		rootContext.Finish()

		fmt.Printf("5 - trying to create new context from finished root context...\n")

		newChildNode := &childNode5{}
		_, err := rootContext.NewContextFor(newChildNode)
		if err != nil {
			_, ok := err.(*context.FinishInProcessForFreezeError)
			if ok {
				fmt.Printf("6 - successfully catched error: %v\n", err)
			} else {
				panic("6 - uncatched error")
			}
		} else {
			panic("6 - uncatched error")
		}
	}()

	fmt.Printf("3 - Wait\n")
	rootContext.Wait()

	fmt.Printf("7 - test finished\n")
}
