package context_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	context "github.com/mcfly722/context"
)

type node5 struct {
	name            string
	sequenceChecker sequenceChecker
	sequenceStep    int
	ready           sync.Mutex
}

func (node *node5) getName() string {
	node.ready.Lock()
	defer node.ready.Unlock()
	return node.name
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
	node.sequenceChecker.NotifyWithText(node.sequenceStep, "%v finished\n", node.getName())
}

func mixedLadder(sequenceChecker sequenceChecker, path string, parents map[context.ChildContext]struct{}, width int, height int) {
	if height > 0 {
		newPath := fmt.Sprintf("%v->%v", path, height)
		newContexts := make(map[context.ChildContext]struct{})

		for i := 0; i < width; i++ {
			newInstance := &node5{
				name:            fmt.Sprintf("%v", newPath),
				sequenceChecker: sequenceChecker,
				sequenceStep:    height,
			}
			fmt.Printf("%v configured\n", newInstance.getName())

			for parent := range parents {
				newContext, _ := parent.NewContextFor(newInstance)
				newContexts[newContext] = struct{}{}
			}
		}

		mixedLadder(sequenceChecker, newPath, newContexts, width, height-1)
	}
}

func Test_MixedLadder(t *testing.T) {
	const ladderHight = 10
	sequenceChecker := newSequenceChecker()

	rootNode := &node5{
		name:            "root",
		sequenceChecker: sequenceChecker,
		sequenceStep:    ladderHight + 1,
	}

	rootContext := context.NewRootContext(rootNode)

	rootContextMap := make(map[context.ChildContext]struct{})
	rootContextMap[rootContext] = struct{}{}

	mixedLadder(sequenceChecker, "root", rootContextMap, 3, ladderHight)

	go func() {
		time.Sleep(100 * time.Millisecond)
		sequenceChecker.NotifyWithText(0, "Close\n")
		rootContext.Close()
	}()

	rootContext.Wait()

	fmt.Printf("finished with correct sequence = %v\n", sequenceChecker.ToString())
}
