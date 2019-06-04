package deepcopy

import (
	"fmt"
	"os"
	"strings"
)

type HierarchyStack string

func (h *HierarchyStack) Push(hierarchy string) {
	if len(*h) == 0 {
		*h = HierarchyStack(hierarchy)
	} else {
		*h = *h + "." + HierarchyStack(hierarchy)
	}
}

func (h *HierarchyStack) Pop() {
	s := string(*h)
	if n := strings.LastIndex(s, "."); n >= 0 {
		*h = HierarchyStack(s[:n])
	} else {
		*h = ""
	}
}

func (h HierarchyStack) Prefix() string {
	return string(h)
}

type Tracer interface {
	Println(args ...interface{})
	PrintfLn(format string, args ...interface{})
}

type stackTracer struct {
	HierarchyStack
	Tracer
}

type traceNothing struct {
}

func (t traceNothing) Println(args ...interface{}) {

}
func (t traceNothing) PrintfLn(format string, args ...interface{}) {

}
func (t traceNothing) Push(hierarchy string) {

}
func (t traceNothing) Pop() {

}
func (t traceNothing) Prefix() string {
	return ""
}

type TraceConsole struct {
	Label string
}

func (t TraceConsole) Println(args ...interface{}) {
	fmt.Fprintln(os.Stderr, append([]interface{}{t.Label}, args...)...)
}

func (t TraceConsole) PrintfLn(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "%s"+format+"\n", append([]interface{}{t.Label}, args...)...)
}
