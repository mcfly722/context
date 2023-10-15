package context_test

import (
	"fmt"
	"testing"
	"time"

	context "github.com/mcfly722/context"
)

type node5 struct {
	name string
}

func newNode5(name string) *node5 {
	return &node5{name: name}
}

func (node *node5) getName() string {
	return node.name
}

func (node *node5) Go(current context.Context[*state5]) {

	fmt.Printf("go:             %v started\n", node.getName())
loop:
	for {
		select {
		case state, isOpened := <-current.Controller():
			if !isOpened {
				break loop
			} else {
				fmt.Printf("2. obtained new state=%v", state.name)
			}
		default:
			{
			}
		}
	}
	fmt.Printf("%v closed\n", node.getName())
}

type state5 struct {
	name string
}

func Test_Send_ToRoot(t *testing.T) {

	rootNode := newNode5("root")

	rootContext := context.NewRootContext[*state5](rootNode)
	fmt.Println("1. root Context created")

	go func() {
		time.Sleep(100 * time.Millisecond)
		fmt.Println("Close")
		rootContext.Close()
	}()

	err := rootContext.Send(&state5{name: "test state"})
	if err != nil {
		t.Fatal(err)
	}

	rootContext.Wait()

	err = rootContext.Send(&state5{name: "failed state test"})
	if err == nil {
		t.Fatal("could not catch send message to closed context")
	} else {
		fmt.Printf("3. successfully catched send error (%v)", err)
	}

	fmt.Println("4. test done")
}
