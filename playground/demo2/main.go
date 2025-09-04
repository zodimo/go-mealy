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
	Input0 mealy.Action = "0" // Represents a '0' input.
	Input1 mealy.Action = "1" // Represents a '1' input.
)

func main() {

	builder := mealy.NewMachineBuilder("Ones tracker")
	builder.AddTransition(mealy.Transition{
		Action:    Input0,
		FromState: StateEven,
		ToState:   StateEven,
	})
	builder.AddTransition(mealy.Transition{
		Action:    Input0,
		FromState: StateOdd,
		ToState:   StateOdd,
	})

	builder.AddTransition(mealy.Transition{
		Action:    Input1,
		FromState: StateEven,
		ToState:   StateOdd,
	})
	builder.AddTransition(mealy.Transition{
		Action:    Input1,
		FromState: StateOdd,
		ToState:   StateEven,
	})
	builder.SetInitialState(StateEven)
	machine, err := builder.Build()
	if err != nil {
		panic(err)
	}

	mealy.WriteMermaidToMarkdownFile(machine, "mealy_diagram.md")
	fmt.Printf("Initial state: %v\n", machine.CurrentState())

	// Simulate a sequence of inputs.
	inputs := []mealy.Action{Input1, Input0, Input1, Input1, Input0, Input1}

	// Process each input and print the results.
	fmt.Println("Processing inputs:", inputs)
	continuation := mealy.NewContinuation(machine)
	for _, input := range inputs {
		currentState := continuation.CurrentState()
		if continuation.GetMachine().CanStep(input) {
			continuation = continuation.GetMachine().StepUnsafe(input)

			fmt.Printf("  Input: %s -> Current State: %v, New State: %v\n", input, currentState, continuation.CurrentState())
		} else {
			panic("Cannot step with input " + string(input))
		}

	}
}
