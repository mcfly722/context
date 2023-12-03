package context_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	context "github.com/mcfly722/context"
)

type node7 struct {
	name            string
	sequenceChecker sequenceChecker
	sequenceStep    int
	ready           sync.Mutex
}

func (node *node7) getName() string {
	node.ready.Lock()
	defer node.ready.Unlock()
	return node.name
}

func (node *node7) Go(current context.Context) {

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

	node.sequenceChecker.Notify(node.sequenceStep)

	fmt.Printf("%v finished\n", node.getName())
}

func (parent *node7) ladder(context context.ChildContext, height int) {
	fmt.Printf("%v configured\n", parent.getName())
	if height > 1 {
		newNode := &node7{
			name:            fmt.Sprintf("%v->%v", parent.getName(), height),
			sequenceChecker: parent.sequenceChecker,
			sequenceStep:    height,
		}
		newContext, err := context.NewContextFor(newNode)
		if err == nil {
			newNode.ladder(newContext, height-1)
		}
	}
}

func Test_Ladder(t *testing.T) {
	const ladderHight = 20
	sequenceChecker := newSequenceChecker()

	rootNode := &node7{
		name:            "root",
		sequenceChecker: sequenceChecker,
		sequenceStep:    ladderHight + 1,
	}

	rootContext := context.NewRootContext(rootNode)

	rootNode.ladder(rootContext, ladderHight)

	go func() {
		time.Sleep(100 * time.Millisecond)
		fmt.Println("Close")
		sequenceChecker.Notify(1)
		rootContext.Close()
	}()

	rootContext.Wait()
	sequenceChecker.Notify(ladderHight + 2)
	fmt.Printf("test done with correct sequence=%v\n", sequenceChecker.ToString())
}
