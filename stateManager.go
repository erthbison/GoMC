package gomc

import (
	"fmt"
	"gomc/event"
	"gomc/tree"
	"io"
	"sync"

	"golang.org/x/exp/maps"
)

type GlobalState[S any] struct {
	LocalStates map[int]S    // A map storing the local state of the nodes. The map stores (id, state) combination.
	Correct     map[int]bool // A map storing the status of the node. The map stores (id, status) combination. If status is true, the node with id "id" is correct, otherwise the node is true. All nodes are represented in the map.
	evt         event.Event  // A copy of the event that caused the transition into this state. It should not be changed.
}

type StateSpace[S any] interface {
	Payload() GlobalState[S]
	Children() []StateSpace[S]
	IsTerminal() bool

	Export(io.Writer)
}

// A wrapper around the Tree structure so that it implements the StateSpace interface
type TreeStateSpace[S any] struct {
	*tree.Tree[GlobalState[S]]
}

func (tss TreeStateSpace[S]) Children() []StateSpace[S] {
	out := []StateSpace[S]{}
	for _, child := range tss.Tree.Children() {
		out = append(out, TreeStateSpace[S]{
			Tree: child,
		})
	}
	return out
}

func (tss TreeStateSpace[S]) IsTerminal() bool {
	return tss.IsLeafNode()
}

func (tss TreeStateSpace[S]) Export(w io.Writer) {
	fmt.Fprint(w, tss.Newick())
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
	AddRun(run []GlobalState[S])
	State() StateSpace[S]
}

// A type that manages the state of several runs of the same system and represent all discovered states of the system.
// The TreeStateManger waits for completed runs on the send channel. Once they are received they are added to the state tree that is used to represent the state space.
// After calling Stop() the send channel will be closed and the TreeStateManager will no longer add new runs to the state tree
type TreeStateManager[T, S any] struct {
	sync.RWMutex
	stateRoot *tree.Tree[GlobalState[S]]

	getLocalState func(*T) S
	stateEq       func(S, S) bool
}

func NewTreeStateManager[T, S any](getLocalState func(*T) S, stateEq func(S, S) bool) *TreeStateManager[T, S] {
	return &TreeStateManager[T, S]{
		getLocalState: getLocalState,
		stateEq:       stateEq,
	}
}

func (sm *TreeStateManager[T, S]) AddRun(run []GlobalState[S]) {
	sm.Lock()
	defer sm.Unlock()
	currentTree := sm.stateRoot
	if len(run) < 1 {
		return
	}
	if currentTree == nil {
		currentTree = sm.initStateTree(run[0])
		sm.stateRoot = currentTree
	}
	for _, state := range run[1:] {
		// If the state already is a child of the current state, retrieve it and set it as the next state
		if nextState := currentTree.GetChild(state); nextState != nil {
			currentTree = nextState
			continue
		}
		// Otherwise add it as a child to the state tree
		currentTree = currentTree.AddChild(state)
	}
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
	return stateRoot
}

// Create a RunStateManager to be used to collect the state of the new run
func (sm *TreeStateManager[T, S]) NewRun() *RunStateManager[T, S] {
	return &RunStateManager[T, S]{
		run:           make([]GlobalState[S], 0),
		getLocalState: sm.getLocalState,
		sm:            sm,
	}
}

// Write the Newick representation of the state tree to the writer
func (sm *TreeStateManager[T, S]) Export(wrt io.Writer) {
	sm.RLock()
	defer sm.RUnlock()
	fmt.Fprint(wrt, sm.stateRoot.Newick())
}

func (sm *TreeStateManager[T, S]) State() StateSpace[S] {
	sm.RLock()
	defer sm.RUnlock()
	return TreeStateSpace[S]{Tree: sm.stateRoot}
}
