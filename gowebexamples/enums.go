package main

import "fmt"

// Our enum type ServerState has an underlying int type
type ServerState int

// The possible values for ServerState are defined as constants.
// The special keyword iota generates successive constant values automatically; in this case 0, 1, 2 and so on.
const (
	StateIddle ServerState = iota
	StateConnected
	StateError
	StateRetrying
)

var stateName = map[ServerState]string{
	StateIddle:     "iddle",
	StateConnected: "connected",
	StateError:     "error",
	StateRetrying:  "retrying",
}

// String() → is a special Go method that defines how the type is printed
// Go’s fmt package checks whether a type implements Stringer interface
// If your type implements that method, then fmt automatically calls it.
func (ss ServerState) String() string {
	return stateName[ss]
}

func main() {
	ns := transition(StateIddle)
	fmt.Println(ns)

	ns2 := transition(ns)
	fmt.Println(ns2)
}

func transition(s ServerState) ServerState {
	switch s {
	case StateIddle:
		return StateConnected
	case StateConnected, StateRetrying:
		return StateIddle
	case StateError:
		return StateError
	default:
		panic(fmt.Errorf("unknow state: %s", s))
	}
}
