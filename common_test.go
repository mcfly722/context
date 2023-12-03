package context_test

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

type sequenceChecker interface {
	Notify(stepNumber uint64)
}

type sequence struct {
	steps []uint64
	ready sync.Mutex
}

func newSequenceChecker() sequenceChecker {
	return &sequence{
		steps: []uint64{},
	}
}

func (current *sequence) Notify(stepNumber uint64) {
	current.ready.Lock()
	defer current.ready.Unlock()

	if len(current.steps) == 0 {
		current.steps = append(current.steps, stepNumber)
		return
	}

	lastStep := current.steps[len(current.steps)-1]
	current.steps = append(current.steps, stepNumber)

	if lastStep > stepNumber {
		panic(fmt.Sprintf("%v incorrect sequence", current.steps))
	}
}

func Test_Sequence1(t *testing.T) {
	sequeceChecker := newSequenceChecker()

	sequeceChecker.Notify(1)
	sequeceChecker.Notify(2)
	sequeceChecker.Notify(3)
	sequeceChecker.Notify(4)
	sequeceChecker.Notify(4)
	sequeceChecker.Notify(5)
}

func mustPanic(f func()) {

	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("successfully catched panic=%v\n", r)
			return
		}
		panic("panic does not catched!")
	}()

	f()
}

func Test_Sequence2(t *testing.T) {
	mustPanic(func() {
		sequeceChecker := newSequenceChecker()
		sequeceChecker.Notify(1)
		sequeceChecker.Notify(2)
		sequeceChecker.Notify(3)
		sequeceChecker.Notify(4)
		sequeceChecker.Notify(3)
	})
}

func Test_SequenceRace(t *testing.T) {

	sequeceChecker := newSequenceChecker()

	for i := 0; i < 100; i++ {
		go func(sequenceChecker sequenceChecker) {
			sequenceChecker.Notify(1)
		}(sequeceChecker)
	}

	time.Sleep(100 * time.Microsecond)
}
