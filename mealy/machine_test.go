package mealy

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestTransition_Validate(t *testing.T) {
	tests := []struct {
		name       string
		transition Transition
		wantErr    bool
		errMsg     string
	}{
		{
			name: "Valid transition",
			transition: Transition{
				Action:    "action",
				FromState: "state1",
				ToState:   "state2",
				Output:    "output",
			},
			wantErr: false,
		},
		{
			name: "Empty action",
			transition: Transition{
				Action:    "",
				FromState: "state1",
				ToState:   "state2",
				Output:    "output",
			},
			wantErr: true,
			errMsg:  "action cannot be empty",
		},
		{
			name: "Empty from state",
			transition: Transition{
				Action:    "action",
				FromState: "",
				ToState:   "state2",
				Output:    "output",
			},
			wantErr: true,
			errMsg:  "from state cannot be empty",
		},
		{
			name: "Empty to state",
			transition: Transition{
				Action:    "action",
				FromState: "state1",
				ToState:   "",
				Output:    "output",
			},
			wantErr: true,
			errMsg:  "to state cannot be empty",
		},
		{
			name: "Empty output",
			transition: Transition{
				Action:    "action",
				FromState: "state1",
				ToState:   "state2",
				Output:    "",
			},
			wantErr: true,
			errMsg:  "output cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.transition.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && err.Error() != tt.errMsg {
				t.Errorf("Validate() error message = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestTransition_CanStep(t *testing.T) {
	transition := Transition{
		Action:    "action1",
		FromState: "state1",
		ToState:   "state2",
		Output:    "output1",
	}

	tests := []struct {
		name      string
		action    Action
		fromState MachineState
		want      bool
	}{
		{
			name:      "Matching action and state",
			action:    "action1",
			fromState: "state1",
			want:      true,
		},
		{
			name:      "Non-matching action",
			action:    "action2",
			fromState: "state1",
			want:      false,
		},
		{
			name:      "Non-matching state",
			action:    "action1",
			fromState: "state3",
			want:      false,
		},
		{
			name:      "Non-matching action and state",
			action:    "action2",
			fromState: "state3",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := transition.CanStep(tt.action, tt.fromState); got != tt.want {
				t.Errorf("CanStep() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Mock observer for testing
type mockObserver struct {
	events []MachineTransitionEvent
}

func (m *mockObserver) OnTransition(event MachineTransitionEvent) {
	m.events = append(m.events, event)
}

func TestNewMachine(t *testing.T) {
	tests := []struct {
		name          string
		machineName   string
		initialState  MachineState
		transitions   []Transition
		wantErr       bool
		errorContains string
	}{
		{
			name:         "Valid machine",
			machineName:  "test-machine",
			initialState: "state1",
			transitions: []Transition{
				{
					Action:    "action1",
					FromState: "state1",
					ToState:   "state2",
					Output:    "output1",
				},
			},
			wantErr: false,
		},
		{
			name:         "Empty machine name",
			machineName:  "",
			initialState: "state1",
			transitions: []Transition{
				{
					Action:    "action1",
					FromState: "state1",
					ToState:   "state2",
					Output:    "output1",
				},
			},
			wantErr:       true,
			errorContains: "machine name cannot be empty",
		},
		{
			name:         "Empty initial state",
			machineName:  "test-machine",
			initialState: "",
			transitions: []Transition{
				{
					Action:    "action1",
					FromState: "state1",
					ToState:   "state2",
					Output:    "output1",
				},
			},
			wantErr:       true,
			errorContains: "initial state cannot be empty",
		},
		{
			name:          "Empty transitions",
			machineName:   "test-machine",
			initialState:  "state1",
			transitions:   []Transition{},
			wantErr:       true,
			errorContains: "transitions cannot be empty",
		},
		{
			name:         "Initial state not in transitions",
			machineName:  "test-machine",
			initialState: "state3",
			transitions: []Transition{
				{
					Action:    "action1",
					FromState: "state1",
					ToState:   "state2",
					Output:    "output1",
				},
			},
			wantErr:       true,
			errorContains: "initial state state3 not found in behavior",
		},
		{
			name:         "Invalid transition",
			machineName:  "test-machine",
			initialState: "state1",
			transitions: []Transition{
				{
					Action:    "",
					FromState: "state1",
					ToState:   "state2",
					Output:    "output1",
				},
			},
			wantErr:       true,
			errorContains: "invalid transition",
		},
		{
			name:         "Duplicate transition",
			machineName:  "test-machine",
			initialState: "state1",
			transitions: []Transition{
				{
					Action:    "action1",
					FromState: "state1",
					ToState:   "state2",
					Output:    "output1",
				},
				{
					Action:    "action1",
					FromState: "state1",
					ToState:   "state3",
					Output:    "output2",
				},
			},
			wantErr:       true,
			errorContains: "duplicate transition",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			machine, err := NewMachine(tt.machineName, tt.initialState, tt.transitions)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewMachine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("NewMachine() error = %v, want to contain %v", err.Error(), tt.errorContains)
				}
				return
			}
			if machine == nil {
				t.Errorf("NewMachine() returned nil machine")
				return
			}
			if machine.GetName() != tt.machineName {
				t.Errorf("Machine name = %v, want %v", machine.GetName(), tt.machineName)
			}
			if machine.CurrentState() != tt.initialState {
				t.Errorf("Initial state = %v, want %v", machine.CurrentState(), tt.initialState)
			}
		})
	}
}

func TestMachine_Step(t *testing.T) {
	// Create a simple machine for testing
	transitions := []Transition{
		{
			Action:    "action1",
			FromState: "state1",
			ToState:   "state2",
			Output:    "output1",
		},
		{
			Action:    "action2",
			FromState: "state2",
			ToState:   "state3",
			Output:    "output2",
		},
		{
			Action:    "action3",
			FromState: "state3",
			ToState:   "state1",
			Output:    "output3",
		},
	}

	observer := &mockObserver{events: []MachineTransitionEvent{}}
	machine, err := NewObservableMachine("test-machine", "state1", transitions, observer)
	if err != nil {
		t.Fatalf("Failed to create machine: %v", err)
	}

	// Test valid step
	output, continuation, err := machine.Step("action1")
	if err != nil {
		t.Errorf("Step() error = %v, wantErr = false", err)
	}
	if output != "output1" {
		t.Errorf("Step() output = %v, want %v", output, "output1")
	}
	if continuation.CurrentState() != "state2" {
		t.Errorf("Step() new state = %v, want %v", continuation.CurrentState(), "state2")
	}

	// Test observer received event
	if len(observer.events) != 1 {
		t.Errorf("Observer events count = %v, want %v", len(observer.events), 1)
	} else {
		event := observer.events[0]
		if event.Action != "action1" || event.FromState != "state1" || event.ToState != "state2" || event.Output != "output1" {
			t.Errorf("Observer event = %+v, want {action1, state1, state2, output1}", event)
		}
	}

	// Test invalid step
	output, continuation, err = machine.Step("invalid-action")
	if !errors.Is(err, ErrNoTransition) {
		t.Errorf("Step() error = %v, want %v", err, ErrNoTransition)
	}
	if output != "" {
		t.Errorf("Step() output = %v, want %v", output, "")
	}
	if continuation != machine {
		t.Errorf("Step() continuation should be the same machine instance")
	}
}

func TestMachine_StepUnsafe(t *testing.T) {
	// Create a simple machine for testing
	transitions := []Transition{
		{
			Action:    "action1",
			FromState: "state1",
			ToState:   "state2",
			Output:    "output1",
		},
	}

	machine, err := NewMachine("test-machine", "state1", transitions)
	if err != nil {
		t.Fatalf("Failed to create machine: %v", err)
	}

	// Test valid step
	output, continuation := machine.StepUnsafe("action1")
	if output != "output1" {
		t.Errorf("StepUnsafe() output = %v, want %v", output, "output1")
	}
	if continuation.CurrentState() != "state2" {
		t.Errorf("StepUnsafe() new state = %v, want %v", continuation.CurrentState(), "state2")
	}

	// Test panic on invalid step
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("StepUnsafe() should panic on invalid action")
		} else if !errors.Is(r.(error), ErrNoTransition) {
			t.Errorf("StepUnsafe() panic = %v, want %v", r, ErrNoTransition)
		}
	}()
	machine.StepUnsafe("invalid-action")
}

func TestMachine_CanStep(t *testing.T) {
	// Create a simple machine for testing
	transitions := []Transition{
		{
			Action:    "action1",
			FromState: "state1",
			ToState:   "state2",
			Output:    "output1",
		},
	}

	machine, err := NewMachine("test-machine", "state1", transitions)
	if err != nil {
		t.Fatalf("Failed to create machine: %v", err)
	}

	// Test valid action
	if !machine.CanStep("action1") {
		t.Errorf("CanStep() = false, want true for valid action")
	}

	// Test invalid action
	if machine.CanStep("invalid-action") {
		t.Errorf("CanStep() = true, want false for invalid action")
	}

	// Step to change state
	machine.Step("action1")

	// Test action that was valid for previous state
	if machine.CanStep("action1") {
		t.Errorf("CanStep() = true, want false for action valid only in previous state")
	}
}

func TestMachine_Reset(t *testing.T) {
	// Create a simple machine for testing
	transitions := []Transition{
		{
			Action:    "action1",
			FromState: "state1",
			ToState:   "state2",
			Output:    "output1",
		},
	}

	machine, err := NewMachine("test-machine", "state1", transitions)
	if err != nil {
		t.Fatalf("Failed to create machine: %v", err)
	}

	// Step to change state
	machine.Step("action1")
	if machine.CurrentState() != "state2" {
		t.Errorf("CurrentState() = %v, want %v after step", machine.CurrentState(), "state2")
	}

	// Reset and check state
	machine.Reset()
	if machine.CurrentState() != "state1" {
		t.Errorf("CurrentState() = %v, want %v after reset", machine.CurrentState(), "state1")
	}
}

func TestMachine_ToMermaid(t *testing.T) {
	// Create a simple machine for testing
	transitions := []Transition{
		{
			Action:    "action1",
			FromState: "state1",
			ToState:   "state2",
			Output:    "output1",
		},
		{
			Action:    "action2",
			FromState: "state2",
			ToState:   "state1",
			Output:    "output2",
		},
	}

	machine, err := NewMachine("test-machine", "state1", transitions)
	if err != nil {
		t.Fatalf("Failed to create machine: %v", err)
	}

	mermaid := machine.ToMermaid()

	// Check for expected elements in the mermaid diagram
	expectedElements := []string{
		"title: test-machine",
		"stateDiagram-v2",
		"[*] --> state1",
		"state1 --> state2 : action1",
		"state2 --> state1 : action2",
	}

	for _, expected := range expectedElements {
		if !strings.Contains(mermaid, expected) {
			t.Errorf("ToMermaid() output doesn't contain expected element: %v", expected)
		}
	}
}

func TestMachine_ToMermaid_MultipleActions(t *testing.T) {
	// Create a machine with multiple actions between the same states
	transitions := []Transition{
		{
			Action:    "action1",
			FromState: "state1",
			ToState:   "state1", // Self-transition
			Output:    "output1",
		},
		{
			Action:    "action2",
			FromState: "state1",
			ToState:   "state1", // Self-transition with different action
			Output:    "output2",
		},
		{
			Action:    "action3",
			FromState: "state1",
			ToState:   "state2",
			Output:    "output3",
		},
		{
			Action:    "action4",
			FromState: "state2",
			ToState:   "state1",
			Output:    "output4",
		},
		{
			Action:    "action5",
			FromState: "state2",
			ToState:   "state1", // Multiple actions to the same transition
			Output:    "output5",
		},
	}

	machine, err := NewMachine("test-multiple-actions", "state1", transitions)
	if err != nil {
		t.Fatalf("Failed to create machine: %v", err)
	}

	mermaid := machine.ToMermaid()

	// Check for expected elements in the mermaid diagram
	expectedElements := []string{
		"title: test-multiple-actions",
		"stateDiagram-v2",
		"[*] --> state1",
		"state1 --> state2 : action3",
	}

	// Check for self-transitions with multiple actions (order doesn't matter)
	selfTransitionLine := "state1 --> state1 :"
	if !strings.Contains(mermaid, selfTransitionLine) {
		t.Errorf("ToMermaid() output doesn't contain self-transition line: %v", selfTransitionLine)
		t.Errorf("Actual mermaid output: %v", mermaid)
	} else {
		// Check that both actions are present in the self-transition line
		selfTransitionContent := extractTransitionContent(mermaid, selfTransitionLine)
		if !strings.Contains(selfTransitionContent, "action1") || !strings.Contains(selfTransitionContent, "action2") {
			t.Errorf("Self-transition line doesn't contain both actions. Line content: %v", selfTransitionContent)
		}
	}

	// Check for state2->state1 transitions with multiple actions (order doesn't matter)
	multiTransitionLine := "state2 --> state1 :"
	if !strings.Contains(mermaid, multiTransitionLine) {
		t.Errorf("ToMermaid() output doesn't contain multi-action transition line: %v", multiTransitionLine)
		t.Errorf("Actual mermaid output: %v", mermaid)
	} else {
		// Check that both actions are present in the multi-action transition line
		multiTransitionContent := extractTransitionContent(mermaid, multiTransitionLine)
		if !strings.Contains(multiTransitionContent, "action4") || !strings.Contains(multiTransitionContent, "action5") {
			t.Errorf("Multi-action transition line doesn't contain both actions. Line content: %v", multiTransitionContent)
		}
	}

	for _, expected := range expectedElements {
		if !strings.Contains(mermaid, expected) {
			t.Errorf("ToMermaid() output doesn't contain expected element: %v", expected)
			t.Errorf("Actual mermaid output: %v", mermaid)
		}
	}
}

// Helper function to extract the content after a transition line prefix
func extractTransitionContent(mermaid, linePrefix string) string {
	lines := strings.Split(mermaid, "\n")
	for _, line := range lines {
		if strings.Contains(line, linePrefix) {
			return strings.TrimSpace(strings.Split(line, ":")[1])
		}
	}
	return ""
}

func TestMachineBuilder(t *testing.T) {
	builder := NewMachineBuilder("test-builder-machine")

	// Test initial state
	builder.SetInitialState("state1")

	// Add transitions
	builder.AddTransition(Transition{
		Action:    "action1",
		FromState: "state1",
		ToState:   "state2",
		Output:    "output1",
	})

	builder.AddTransition(Transition{
		Action:    "action2",
		FromState: "state2",
		ToState:   "state1",
		Output:    "output2",
	})

	// Build the machine
	machine, err := builder.Build()
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	// Verify machine properties
	if machine.GetName() != "test-builder-machine" {
		t.Errorf("Machine name = %v, want %v", machine.GetName(), "test-builder-machine")
	}

	if machine.CurrentState() != "state1" {
		t.Errorf("Initial state = %v, want %v", machine.CurrentState(), "state1")
	}

	// Test machine functionality
	if !machine.CanStep("action1") {
		t.Errorf("CanStep() = false, want true for action1")
	}

	output, continuation, err := machine.Step("action1")
	if err != nil {
		t.Errorf("Step() error = %v", err)
	}
	if output != "output1" {
		t.Errorf("Step() output = %v, want %v", output, "output1")
	}
	if continuation.CurrentState() != "state2" {
		t.Errorf("Step() new state = %v, want %v", continuation.CurrentState(), "state2")
	}
}

func TestContinuation(t *testing.T) {
	// Create a simple machine for testing
	transitions := []Transition{
		{
			Action:    "action1",
			FromState: "state1",
			ToState:   "state2",
			Output:    "output1",
		},
	}

	machine, err := NewMachine("test-machine", "state1", transitions)
	if err != nil {
		t.Fatalf("Failed to create machine: %v", err)
	}

	// Create continuation
	continuation := NewContinuation(machine)

	// Test continuation methods
	if continuation.CurrentState() != "state1" {
		t.Errorf("CurrentState() = %v, want %v", continuation.CurrentState(), "state1")
	}

	if !reflect.DeepEqual(continuation.GetMachine(), machine) {
		t.Errorf("GetMachine() returned different machine instance")
	}

	// Step the machine through continuation
	output, newContinuation, err := continuation.GetMachine().Step("action1")
	if err != nil {
		t.Errorf("Step() through continuation error = %v", err)
	}
	if output != "output1" {
		t.Errorf("Step() through continuation output = %v, want %v", output, "output1")
	}
	if newContinuation.CurrentState() != "state2" {
		t.Errorf("Step() through continuation new state = %v, want %v", newContinuation.CurrentState(), "state2")
	}
}

func TestBuildBehavior(t *testing.T) {
	transitions := []Transition{
		{
			Action:    "action1",
			FromState: "state1",
			ToState:   "state2",
			Output:    "output1",
		},
		{
			Action:    "action2",
			FromState: "state1",
			ToState:   "state3",
			Output:    "output2",
		},
		{
			Action:    "action3",
			FromState: "state2",
			ToState:   "state1",
			Output:    "output3",
		},
	}

	behavior, err := buildBehavior(transitions)
	if err != nil {
		t.Fatalf("buildBehavior() error = %v", err)
	}

	// Check behavior structure
	if len(behavior) != 2 {
		t.Errorf("behavior has %v states, want %v", len(behavior), 2)
	}

	// Check state1 transitions
	if len(behavior["state1"]) != 2 {
		t.Errorf("state1 has %v actions, want %v", len(behavior["state1"]), 2)
	}

	// Check specific transitions
	if behavior["state1"]["action1"].ToState != "state2" {
		t.Errorf("state1->action1 goes to %v, want %v", behavior["state1"]["action1"].ToState, "state2")
	}

	if behavior["state1"]["action2"].ToState != "state3" {
		t.Errorf("state1->action2 goes to %v, want %v", behavior["state1"]["action2"].ToState, "state3")
	}

	if behavior["state2"]["action3"].ToState != "state1" {
		t.Errorf("state2->action3 goes to %v, want %v", behavior["state2"]["action3"].ToState, "state1")
	}

	// Test invalid behavior
	invalidTransitions := []Transition{
		{
			Action:    "action1",
			FromState: "state1",
			ToState:   "state2",
			Output:    "output1",
		},
		{
			Action:    "action1", // Duplicate action for state1
			FromState: "state1",
			ToState:   "state3",
			Output:    "output2",
		},
	}

	_, err = buildBehavior(invalidTransitions)
	if err == nil {
		t.Errorf("buildBehavior() with duplicate transitions should return error")
	} else if !strings.Contains(err.Error(), "duplicate transition") {
		t.Errorf("buildBehavior() error = %v, want error containing 'duplicate transition'", err)
	}
}
