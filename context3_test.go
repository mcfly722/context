package context_test

import (
	"fmt"
	"testing"

	context "github.com/mcfly722/context"
)

type node3 struct{}

func (node *node3) Go(current context.Context) {
	current.SetDefer(func(recover interface{}) {

		fmt.Printf("custom defer runned\n")
		if recover != nil {
			fmt.Printf("successfully catched Panic: %v\n", recover)
		}

	})

	fmt.Printf("defer configured\n")

	// exit without Cancel() panic ...
}

func Test_DeferHandler_WithoutCancel(t *testing.T) {
	rootNode := &node3{}

	rootContext, err := context.NewRootContext(rootNode)
	if err != nil {
		t.Fatal(err)
	}

	rootContext.Wait()
	fmt.Printf("test finished\n")
}
