package context_test

import (
	"fmt"
	"testing"

	context "github.com/mcfly722/context"
)

type node3 struct{}

func (node *node3) Go(current context.Context) {

	current.SetDefer(func(recover interface{}) {

		fmt.Printf("defer:       custom defer runned\n")
		if recover != nil {
			if recover == context.ExitFromContextWithoutCancelPanic {
				fmt.Printf("defer:       successfully catched Panic: %v\n", recover)
			}
		}

	})

	fmt.Printf("go:          defer configured\n")

	// exit without Cancel() panic ...
}

func Test_DeferHandler_WithoutCancel(t *testing.T) {
	rootNode := &node3{}

	fmt.Printf("1. starting new root context\n")
	rootContext := context.NewRootContext(rootNode)

	fmt.Printf("2. wait\n")
	rootContext.Wait()
	fmt.Printf("3. test finished\n")
}
