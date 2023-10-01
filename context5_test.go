package context_test

import (
	"fmt"
	"testing"
	"time"

	context "github.com/mcfly722/context"
)

type childNode5 struct{}
type rootNode5 struct{}

func (node *childNode5) Go(current context.Context) {
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
	fmt.Printf("go:     childNode disposing for 300ms\n")
	time.Sleep(300 * time.Millisecond)
	fmt.Printf("go:     childNode finished\n")
}

func (node *rootNode5) Go(current context.Context) {
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
	fmt.Printf("go:     rootNode finished\n")
}

func Test_FailCreateContextFromRootNode(t *testing.T) {

	rootNode := &rootNode5{}
	childNode := &childNode5{}

	fmt.Printf("1 - creating new root context\n")
	rootContext, err := context.NewRootContext(rootNode)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("2 - creating child context\n")
	_, err = rootContext.NewContextFor(childNode)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		time.Sleep(100 * time.Millisecond)
		fmt.Printf("4 - Canceling root context\n")
		rootContext.Cancel()

		fmt.Printf("5 - trying to create new context from closed root context...\n")

		newChildNode := &childNode5{}
		_, err := rootContext.NewContextFor(newChildNode)
		if err != nil {
			_, ok := err.(*context.CancelInProcessForFreezeError)
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
