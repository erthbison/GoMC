package gomc

import (
	"fmt"
	"gomc/event"
	"gomc/tree"
	"io"

	"golang.org/x/exp/maps"
)

type GlobalState[S any] struct {
	LocalStates map[int]S    // A map storing the local state of the nodes. The map stores (id, state) combination.
	Correct     map[int]bool // A map storing the status of the node. The map stores (id, status) combination. If status is true, the node with id "id" is correct, otherwise the node is true. All nodes are represented in the map.
	evt         event.Event  // A copy of the event that caused the transition into this state. It should not be changed.
}

func (gs GlobalState[S]) String() string {
	crashed := []int{}
	for id, status := range gs.Correct {
		if !status {
			crashed = append(crashed, id)
		}
	}
	return fmt.Sprintf("Evt: %v\t States: %v\t Crashed: %v\t", gs.evt, gs.LocalStates, crashed)
}

// Manages the global state across several runs.
type StateManager[T, S any] interface {
	NewRun() *RunStateManager[T, S]
}

// A type that manages the state of several runs of the same system and represent all discovered states of the system.
// The TreeStateManger waits for completed runs on the send channel. Once they are received they are added to the state tree that is used to represent the state space.
// After calling Stop() the send channel will be closed and the TreeStateManager will no longer add new runs to the state tree
type TreeStateManager[T, S any] struct {
	StateRoot *tree.Tree[GlobalState[S]]

	getLocalState func(*T) S
	stateEq       func(S, S) bool
	send          chan []GlobalState[S]
}

func NewTreeStateManager[T, S any](getLocalState func(*T) S, stateEq func(S, S) bool) *TreeStateManager[T, S] {
	sm := &TreeStateManager[T, S]{
		getLocalState: getLocalState,
		stateEq:       stateEq,
		send:          make(chan []GlobalState[S]),
	}
	go sm.start()
	return sm
}

func (sm *TreeStateManager[T, S]) start() {
	for run := range sm.send {
		if run == nil {
			close(sm.send)
		}
		currentTree := sm.StateRoot
		if len(run) < 1 {
			continue
		}
		if currentTree == nil {
			currentTree = sm.initStateTree(run[0])
		}
		for _, state := range run[1:] {

			// If the state already is a child of the current state, retrieve it and set it as the next state
			if nextState := currentTree.GetChild(state); nextState != nil {
				currentTree = nextState
				continue
			}
			//  Otherwise add it as a child to the state tree
			currentTree = currentTree.AddChild(state)
		}
	}
}

func (sm *TreeStateManager[T, S]) Stop() {
	sm.send <- nil
}

// Initializes the state tree with the provided state as the initial state
func (sm *TreeStateManager[T, S]) initStateTree(state GlobalState[S]) *tree.Tree[GlobalState[S]] {
	cmp := func(a, b GlobalState[S]) bool {
		if !event.EventsEquals(a.evt, b.evt) {
			return false
		}
		if !maps.EqualFunc(a.LocalStates, b.LocalStates, sm.stateEq) {
			return false
		}
		return maps.Equal(a.Correct, b.Correct)
	}
	stateRoot := tree.New(state, cmp)
	sm.StateRoot = &stateRoot
	return &stateRoot
}

// Create a RunStateManager to be used to collect the state of the new run
func (sm *TreeStateManager[T, S]) NewRun() *RunStateManager[T, S] {
	return &RunStateManager[T, S]{
		run:           make([]GlobalState[S], 0),
		getLocalState: sm.getLocalState,
		send:          sm.send,
	}
}

// Write the Newick representation of the state tree to the writer
func (sm *TreeStateManager[T, S]) Export(wrt io.Writer) {
	fmt.Fprint(wrt, sm.StateRoot.Newick())
}
