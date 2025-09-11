package main

import (
	"fmt"

	"github.com/zodimo/go-mealy/mealy"
)

// Define the possible states of our machine.

const (
	StateEven mealy.MachineState = "even" // The state for an even number of '1's seen so far.
	StateOdd  mealy.MachineState = "odd"  // The state for an odd number of '1's seen so far.
)

const (
	OutputEven mealy.Output = "even"
	OutputOdd  mealy.Output = "odd"
)

const (
	Input0 mealy.Action = "0" // Represents a '0' input.
	Input1 mealy.Action = "1" // Represents a '1' input.
	Input2 mealy.Action = "2" // Additional input to demonstrate multiple actions
	Input3 mealy.Action = "3" // Additional input to demonstrate multiple actions
)

func main() {

	builder := mealy.NewMachineBuilder("Ones tracker")
	// Original transitions
	builder.AddTransition(mealy.Transition{
		Action:    Input0,
		FromState: StateEven,
		ToState:   StateEven,
		Output:    OutputEven,
	})
	builder.AddTransition(mealy.Transition{
		Action:    Input0,
		FromState: StateOdd,
		ToState:   StateOdd,
		Output:    OutputOdd,
	})

	builder.AddTransition(mealy.Transition{
		Action:    Input1,
		FromState: StateEven,
		ToState:   StateOdd,
		Output:    OutputOdd,
	})
	builder.AddTransition(mealy.Transition{
		Action:    Input1,
		FromState: StateOdd,
		ToState:   StateEven,
		Output:    OutputEven,
	})

	// Additional transitions to demonstrate multiple actions to the same state
	builder.AddTransition(mealy.Transition{
		Action:    Input2,
		FromState: StateEven,
		ToState:   StateEven, // Same from->to state as Input0
		Output:    OutputEven,
	})
	builder.AddTransition(mealy.Transition{
		Action:    Input3,
		FromState: StateEven,
		ToState:   StateEven, // Same from->to state as Input0 and Input2
		Output:    OutputEven,
	})

	builder.SetInitialState(StateEven)
	machine, err := builder.Build()
	if err != nil {
		panic(err)
	}

	mealy.WriteMermaidToMarkdownFile(machine, "mealy_diagram.md")
	fmt.Printf("Initial state: %v\n", machine.CurrentState())

	// Simulate a sequence of inputs.
	inputs := []mealy.Action{Input1, Input0, Input1, Input1, Input0, Input1, Input2, Input3}

	// Process each input and print the results.
	fmt.Println("Processing inputs:", inputs)
	continuation := mealy.NewContinuation(machine)
	for _, input := range inputs {
		if continuation.GetMachine().CanStep(input) {
			output, continuation := continuation.GetMachine().StepUnsafe(input)
			fmt.Printf("  Input: %s -> New State: %v, Output: %v\n", input, continuation.CurrentState(), output)
		} else {
			panic("Cannot step with input " + string(input))
		}
	}
}
