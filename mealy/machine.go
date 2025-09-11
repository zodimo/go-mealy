package mealy

import (
	"fmt"
	"os"
	"strings"
	"sync"
)

type MachineState string
type Action string
type Output string

type MachineTransitionEvent struct {
	Action    Action
	FromState MachineState
	ToState   MachineState
	Output    Output
}

type MachineObserver interface {
	//action, state , new state output
	OnTransition(event MachineTransitionEvent)
}

var _ MachineObserver = (*noopObserver)(nil)

type noopObserver struct {
}

func (o *noopObserver) OnTransition(event MachineTransitionEvent) {
	// noop
}

type Machine interface {
	Continuation
	Reset()
	Step(input Action) (output Output, continuation Continuation, err error)
	StepUnsafe(input Action) (output Output, continuation Continuation)
	CanStep(input Action) bool
	ToMermaid() string
	GetName() string
}

// MealyMachine represents a Mealy machine with states, transitions, and outputs.
// the output is the current state

type WithCurrentState interface {
	CurrentState() MachineState
}

type WithMachine interface {
	GetMachine() Machine
}

type Continuation interface {
	WithCurrentState
	WithMachine
}

// action + state
type Transition struct {
	Action    Action
	FromState MachineState
	ToState   MachineState
	Output    Output
}

func (t Transition) Validate() error {
	if t.Action == "" {
		return fmt.Errorf("action cannot be empty")
	}
	if t.FromState == "" {
		return fmt.Errorf("from state cannot be empty")
	}
	if t.ToState == "" {
		return fmt.Errorf("to state cannot be empty")
	}
	if t.Output == "" {
		return fmt.Errorf("output cannot be empty")
	}
	return nil
}

func (t Transition) CanStep(action Action, fromState MachineState) bool {
	return t.Action == action && t.FromState == fromState
}

type continuation struct {
	machine Machine
}

var ErrNoTransition = fmt.Errorf("no valid transition found")

var _ Machine = (*machine)(nil)

type machine struct {
	name         string
	currentState MachineState
	behavior     Behavior
	initialState MachineState
	observer     MachineObserver
	mutex        sync.Mutex
}

func (m *machine) Reset() {
	m.currentState = m.initialState
}

func (m *machine) Step(input Action) (output Output, continuation Continuation, err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if transitions, ok := m.behavior[m.currentState]; ok {
		if t, ok := transitions[input]; ok {

			m.currentState = t.ToState
			m.observer.OnTransition(MachineTransitionEvent{
				Action:    input,
				FromState: t.FromState,
				ToState:   t.ToState,
				Output:    t.Output,
			})
			return t.Output, NewContinuation(m), nil
		}
	}
	return "", m, ErrNoTransition
}
func (m *machine) StepUnsafe(input Action) (output Output, continuation Continuation) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if transitions, ok := m.behavior[m.currentState]; ok {
		if t, ok := transitions[input]; ok {

			m.currentState = t.ToState
			m.observer.OnTransition(MachineTransitionEvent{
				Action:    input,
				FromState: t.FromState,
				ToState:   t.ToState,
				Output:    t.Output,
			})
			return t.Output, NewContinuation(m)
		}
	}
	panic(ErrNoTransition)
}

func (m *machine) CanStep(input Action) bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if transitions, ok := m.behavior[m.currentState]; ok {
		if _, ok := transitions[input]; ok {
			return true
		}
	}
	return false
}

func (m *machine) CurrentState() MachineState {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.currentState
}
func (m *machine) GetMachine() Machine {
	return m
}
func (m *machine) GetName() string {
	return m.name
}
func (m *machine) GetOutput() Output {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return Output(m.currentState)
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

func NewObservableMachine(name string, initialState MachineState, transitions []Transition, observer MachineObserver) (Machine, error) {
	if name == "" {
		return nil, fmt.Errorf("machine name cannot be empty")
	}

	if string(initialState) == "" {
		return nil, fmt.Errorf("initial state cannot be empty")
	}

	if len(transitions) == 0 {
		return nil, fmt.Errorf("transitions cannot be empty")
	}
	behavior, err := buildBehavior(transitions)
	if err != nil {
		return nil, err
	}

	if _, ok := behavior[initialState]; !ok {
		return nil, fmt.Errorf("initial state %s not found in behavior", initialState)
	}
	return &machine{
		name:         name,
		currentState: initialState,
		initialState: initialState,
		behavior:     behavior,
		observer:     observer,
	}, nil
}

func NewMachine(name string, initialState MachineState, transitions []Transition) (Machine, error) {
	return NewObservableMachine(name, initialState, transitions, &noopObserver{})
}

// Machine builder
type MachineBuilder struct {
	name         string
	initialState MachineState
	transitions  []Transition
}

func NewMachineBuilder(name string) *MachineBuilder {
	return &MachineBuilder{
		name: name,
	}
}
func (mb *MachineBuilder) AddTransition(t Transition) *MachineBuilder {
	mb.transitions = append(mb.transitions, t)
	return mb
}

func (mb *MachineBuilder) SetInitialState(initialState MachineState) *MachineBuilder {
	mb.initialState = initialState
	return mb
}

func (mb *MachineBuilder) Build() (Machine, error) {
	return NewMachine(mb.name, mb.initialState, mb.transitions)
}

type Behavior map[MachineState]map[Action]Transition

func buildBehavior(transitions []Transition) (Behavior, error) {
	behavior := make(Behavior)
	for _, t := range transitions {
		if err := t.Validate(); err != nil {
			return nil, fmt.Errorf("invalid transition: %w", err)
		}
		// check for duplicate transitions
		if _, ok := behavior[t.FromState]; ok {
			if _, ok := behavior[t.FromState][t.Action]; ok {
				return nil, fmt.Errorf("duplicate transition for action %s from state %s", t.Action, t.FromState)
			}
		}
		if behavior[t.FromState] == nil {
			behavior[t.FromState] = make(map[Action]Transition)
		}
		behavior[t.FromState][t.Action] = t
	}
	return behavior, nil
}

func (m *machine) ToMermaid() string {

	titleString := fmt.Sprintf("---\ntitle: %s\n---\n", m.GetName())

	result := fmt.Sprintf("%s stateDiagram-v2\n", titleString)

	result += fmt.Sprintf("    [*] --> %s\n", m.initialState)

	// Group transitions by from-state and to-state
	transitionMap := make(map[string]map[string][]string) // fromState -> toState -> []actions with outputs

	// Add states and transitions
	for fromState, actions := range m.behavior {
		for action, transition := range actions {
			fromStateStr := string(fromState)
			toStateStr := string(transition.ToState)

			// Initialize maps if they don't exist
			if transitionMap[fromStateStr] == nil {
				transitionMap[fromStateStr] = make(map[string][]string)
			}

			// Add action with output to the appropriate transition group
			transitionMap[fromStateStr][toStateStr] = append(
				transitionMap[fromStateStr][toStateStr],
				fmt.Sprintf("%s -> %s", action, transition.Output),
			)
		}
	}

	// Generate diagram with grouped actions
	for fromState, toStates := range transitionMap {
		for toState, actions := range toStates {
			// Join all actions with a comma and space
			actionsStr := strings.Join(actions, ", ")
			result += fmt.Sprintf("    %s --> %s : %s\n", fromState, toState, actionsStr)
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
