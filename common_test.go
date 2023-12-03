package context_test

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

type sequenceChecker interface {
	Notify(stepNumber int)
	NotifyWithText(stepNumber int, msg string, a ...interface{})
	ToString() string
}

type sequence struct {
	steps []int
	ready sync.Mutex
}

func newSequenceChecker() sequenceChecker {
	return &sequence{
		steps: []int{},
	}
}

func (current *sequence) Notify(stepNumber int) {
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

func (current *sequence) NotifyWithText(stepNumber int, msg string, a ...interface{}) {
	current.Notify(stepNumber)
	fmt.Printf("%v - ", stepNumber)
	fmt.Printf(msg, a...)
}

func (current *sequence) ToString() string {
	current.ready.Lock()
	defer current.ready.Unlock()
	return fmt.Sprintf("%v", current.steps)
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

	for i := 0; i < 30; i++ {
		go func(sequenceChecker sequenceChecker) {
			sequenceChecker.Notify(0)
			fmt.Printf("%v\n", sequenceChecker.ToString())
		}(sequeceChecker)
	}

	time.Sleep(100 * time.Microsecond)
}
