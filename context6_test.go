package context_test

import (
	"fmt"
	"testing"
	"time"

	context "github.com/mcfly722/context"
)

type node6 struct {
	name            string
	lifeTimeMS      uint64
	sequenceChecker sequenceChecker
	sequenceStep    int
}

func (node *node6) Go(current context.Context) {
	fmt.Printf("       %v started\n", node.name)
loop:
	for {
		select {
		case <-time.After(time.Millisecond * time.Duration(node.lifeTimeMS)):
			break loop
		case _, isOpened := <-current.Context():
			if !isOpened {
				break loop
			}
		}
	}
	fmt.Printf("%v finished\n", node.name)
	node.sequenceChecker.Notify(node.sequenceStep)

}

const dynamicPoolSize = 15

func Test_DynamicPool(t *testing.T) {
	sequenceChecker := newSequenceChecker()

	rootNode := &node6{
		name:            "root",
		lifeTimeMS:      1000000,
		sequenceChecker: sequenceChecker,
		sequenceStep:    2,
	}

	inputNode := &node6{
		name:            "input",
		lifeTimeMS:      1000000,
		sequenceChecker: sequenceChecker,
		sequenceStep:    1,
	}

	rootContext := context.NewRootContext(rootNode)

	for i := 0; i < dynamicPoolSize; i++ {
		workerNode := &node6{
			name:            fmt.Sprintf("worker[%v]", i),
			lifeTimeMS:      uint64(100*dynamicPoolSize - 100*i),
			sequenceChecker: sequenceChecker,
			sequenceStep:    -i,
		}
		fmt.Printf("%v configured\n", workerNode.name)

		workerContext, err := rootContext.NewContextFor(workerNode)
		if err != nil {
			t.Fatal(err)
		}

		_, err = workerContext.NewContextFor(inputNode)
		if err != nil {
			t.Fatal(err)
		}

	}

	go func() {
		time.Sleep(2 * dynamicPoolSize * 100 * time.Millisecond)
		rootContext.Close()
	}()

	rootContext.Wait()

	fmt.Printf("finished with correct sequence = %v\n", sequenceChecker.ToString())
}
