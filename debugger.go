package context

import (
	"fmt"
	"strings"
	"time"
)

// DebugNode ...
type DebugNode struct {
	ID            int64
	ComponentType string
	ComponentName string
}

// Debugger ...
type Debugger interface {
	Log(nodePath []DebugNode, eventType int, msg string)
}

// EmptyDebugger ...
type EmptyDebugger struct{}

// Log ...
func (emptyDebugger *EmptyDebugger) Log(nodePath []DebugNode, eventType int, msg string) {}

// NewEmptyDebugger ...
func NewEmptyDebugger() Debugger {
	return &EmptyDebugger{}
}

// ConsoleLogDebugger ...
type ConsoleLogDebugger struct{}

// NewConsoleLogDebugger ...
func NewConsoleLogDebugger() Debugger {
	return &ConsoleLogDebugger{}
}

// Log ...
func (consoleLogDebugger *ConsoleLogDebugger) Log(nodePath []DebugNode, eventType int, msg string) {
	pathStrings := []string{}

	for _, node := range nodePath {
		pathStrings = append(pathStrings, node.ComponentName)
	}

	path := strings.Join(pathStrings, "->")

	fmt.Println(fmt.Sprintf("%v %4v [%v]: %v", time.Now().Format(time.RFC3339), eventType, path, msg))
}
