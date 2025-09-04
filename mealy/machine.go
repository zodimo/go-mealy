package mealy

import (
	"fmt"
	"os"
)

type MachineState string
type Action string

type Machine interface {
	Continuation
	Reset()
	Step(input Action) (output Continuation, err error)
	StepUnsafe(input Action) Continuation
	CanStep(input Action) bool
	ToMermaid() string
}

// MealyMachine represents a Mealy machine with states, transitions, and outputs.
// the output is the current state

type WithCurrentState interface {
	CurrentState() MachineState
}

type WithMachine interface {
	GetMachine() Machine
}

// or Tuple of both
type Continuation interface {
	WithCurrentState
	WithMachine
}

// action + state
type Transition struct {
	Action    Action
	FromState MachineState
	ToState   MachineState
}

func (t Transition) CanStep(action Action, fromState MachineState) bool {
	return t.Action == action && t.FromState == fromState
}

func (t Transition) Step() MachineState {
	return t.ToState
}

type continuation struct {
	machine Machine
}

var ErrNoTransition = fmt.Errorf("no valid transition found")

var _ Machine = (*machine)(nil)

type machine struct {
	currentState MachineState
	behavior     Behavior
	initialState MachineState
}

func (m *machine) Reset() {
	m.currentState = m.initialState
}

func (m *machine) Step(input Action) (output Continuation, err error) {
	if transitions, ok := m.behavior[m.currentState]; ok {
		if t, ok := transitions[input]; ok {

			m.currentState = t.Step()
			return NewContinuation(m), nil
		}
	}
	return nil, ErrNoTransition
}
func (m *machine) StepUnsafe(input Action) Continuation {
	if transitions, ok := m.behavior[m.currentState]; ok {
		if t, ok := transitions[input]; ok {

			m.currentState = t.Step()
			return NewContinuation(m)
		}
	}
	panic(ErrNoTransition)
}

func (m *machine) CanStep(input Action) bool {
	if transitions, ok := m.behavior[m.currentState]; ok {
		if _, ok := transitions[input]; ok {
			return true
		}
	}
	return false
}

func (m *machine) CurrentState() MachineState {
	return m.currentState
}
func (m *machine) GetMachine() Machine {
	return m
}

func (c continuation) CurrentState() MachineState {
	return c.machine.(WithCurrentState).CurrentState()
}
func (c continuation) GetMachine() Machine {
	return c.machine
}

func NewContinuation(m Machine) Continuation {
	return continuation{machine: m}
}

func NewMachine(initialState MachineState, transitions []Transition) Machine {
	return &machine{
		currentState: initialState,
		initialState: initialState,
		behavior:     buildBehavior(transitions),
	}
}

// Machine builder
type MachineBuilder struct {
	transitions []Transition
}

func NewMachineBuilder() *MachineBuilder {
	return &MachineBuilder{}
}
func (mb *MachineBuilder) AddTransition(t Transition) *MachineBuilder {
	mb.transitions = append(mb.transitions, t)
	return mb
}

func (mb *MachineBuilder) Build(initialState MachineState) Machine {
	return NewMachine(initialState, mb.transitions)
}

type Behavior map[MachineState]map[Action]Transition

func buildBehavior(transitions []Transition) Behavior {
	behavior := make(Behavior)
	for _, t := range transitions {
		if behavior[t.FromState] == nil {
			behavior[t.FromState] = make(map[Action]Transition)
		}
		behavior[t.FromState][t.Action] = t
	}
	return behavior
}

func (m *machine) ToMermaid() string {
	result := "stateDiagram-v2\n"
	// Add states and transitions
	for fromState, actions := range m.behavior {
		for action, transition := range actions {
			result += fmt.Sprintf("    %s --> %s : %s\n", fromState, transition.ToState, action)
		}
	}
	return result
}

func WriteMermaidToMarkdownFile(m Machine, filename string) error {
	content := m.(*machine).ToMermaid()
	markdown := fmt.Sprintf("```mermaid\n%s\n```", content)
	return writeToFile(filename, markdown)
}

func writeToFile(filename, content string) error {
	return os.WriteFile(filename, []byte(content), 0644)
}
