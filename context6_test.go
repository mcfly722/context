package context_test

import (
	"fmt"
	"sync"
	"testing"

	context "github.com/mcfly722/context"
)

type node6 struct {
	name string
}

func newNode6(name string) *node6 {
	return &node6{name: name}
}

type state6 struct{}

var messages6 int = 0
var messagesLock6 sync.Mutex

func (node *node6) Go(current context.Context[state6]) {
loop:
	for {
		select {
		case _, isOpened := <-current.Controller():
			if !isOpened {
				break loop
			} else {
				messagesLock6.Lock()
				messages6++
				messagesLock6.Unlock()
			}
		default:
			{
			}
		}
	}
}

func Test_Send_Close_Root_Race(t *testing.T) {

	rootNode := newNode6("root")

	for i := 0; i < 10000; i++ {

		if (i % 100) == 0 {
			fmt.Printf("Test_Send_Close_Root_Race %v/%v messages=%v\n", i, race_iterations, messages6)
		}

		rootContext := context.NewRootContext[state6](rootNode)

		go func() {
			rootContext.Send(state6{})
			rootContext.Close()
		}()

		go func() {
			rootContext.Close()
		}()

		rootContext.Wait()

	}

}
