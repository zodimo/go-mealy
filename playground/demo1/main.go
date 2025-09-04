package main

import "fmt"

// Mealy State Machine is a type of state machine where the output is determined by
// both the current state and the current input. This is in contrast to a Moore machine,
// where the output is determined solely by the current state.

// Define the possible states of our machine.
type State int

const (
	StateEven State = iota // The state for an even number of '1's seen so far.
	StateOdd               // The state for an odd number of '1's seen so far.
)

// Define the possible inputs to our machine.
type Input int

const (
	Input0 Input = 0 // Represents a '0' input.
	Input1 Input = 1 // Represents a '1' input.
)

// The Output of the machine. For this example, the output is a simple string.
type Output string

const (
	OutputEven Output = "even" // Output when an even number of '1's have been seen.
	OutputOdd  Output = "odd"  // Output when an odd number of '1's have been seen.
)

// Transition defines the next state and output for a given current state and input.
type Transition struct {
	NextState State
	Output    Output
}

// MealyMachine represents the state machine itself.
type MealyMachine struct {
	CurrentState State
	// The core of the Mealy machine: a map representing all possible transitions.
	// The outer map key is the current state.
	// The inner map key is the input.
	// The value is the Transition struct (next state and output).
	transitions map[State]map[Input]Transition
}

// NewMealyMachine creates and initializes a new Mealy machine.
func NewMealyMachine() *MealyMachine {
	m := &MealyMachine{
		CurrentState: StateEven, // Start in the 'even' state.
		transitions:  make(map[State]map[Input]Transition),
	}

	// Define the transition logic for the 'even' state.
	m.transitions[StateEven] = map[Input]Transition{
		Input0: {NextState: StateEven, Output: OutputEven}, // From Even, input 0 -> stay Even, output Even.
		Input1: {NextState: StateOdd, Output: OutputOdd},   // From Even, input 1 -> go to Odd, output Odd.
	}

	// Define the transition logic for the 'odd' state.
	m.transitions[StateOdd] = map[Input]Transition{
		Input0: {NextState: StateOdd, Output: OutputOdd},   // From Odd, input 0 -> stay Odd, output Odd.
		Input1: {NextState: StateEven, Output: OutputEven}, // From Odd, input 1 -> go to Even, output Even.
	}

	return m
}

// ProcessInput takes an input, updates the machine's state, and returns the output.
func (m *MealyMachine) ProcessInput(input Input) Output {
	// Check if a transition exists for the current state and input.
	if transitionsForState, ok := m.transitions[m.CurrentState]; ok {
		if transition, ok := transitionsForState[input]; ok {
			// Update the current state to the next state.
			m.CurrentState = transition.NextState
			// Return the output associated with this transition.
			return transition.Output
		}
	}

	// Handle the case where a transition is not defined.
	// In a real application, you would handle this error more gracefully.
	panic("Invalid state or input transition")
}

func main() {
	// Create a new Mealy machine instance.
	machine := NewMealyMachine()
	fmt.Printf("Initial state: %v\n", machine.CurrentState)

	// Simulate a sequence of inputs.
	inputs := []Input{Input1, Input0, Input1, Input1, Input0, Input1}

	// Process each input and print the results.
	fmt.Println("Processing inputs:", inputs)
	for _, input := range inputs {
		output := machine.ProcessInput(input)
		fmt.Printf("  Input: %d -> New State: %v, Output: %v\n", input, machine.CurrentState, output)
	}
}
