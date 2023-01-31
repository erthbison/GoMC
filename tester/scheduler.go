package tester

import (
	"experimentation/sequence"
	"experimentation/tree"
)

type Scheduler interface {
	GetEvent() (*Event, error)
	AddEvent(Event)
	EndRun()
}

type BasicScheduler struct {
	EventRoot    tree.Tree[Event]
	currentEvent *tree.Tree[Event]
}

func NewBasicScheduler() *BasicScheduler {
	eventTree := tree.New(Event{Type: "Start"}, EventsEquals)
	return &BasicScheduler{
		EventRoot:    eventTree,
		currentEvent: &eventTree,
	}
}

func (bs *BasicScheduler) GetEvent() (*Event, error) {
	for _, child := range bs.currentEvent.Children {
		// iteratively check if each child can be the next event
		// a child can be the next event if it has some descendent leaf node that is not an "End" event
		if child.SearchLeafNodes(func(e Event) bool { return e.Type != "End" }) {
			bs.currentEvent = child
			return &child.Payload, nil
		}
	}
	return nil, NoEventError
}

func (bs *BasicScheduler) AddEvent(event Event) {
	if !bs.currentEvent.HasChild(event) {
		bs.currentEvent.AddChild(event)
	}
}

func (bs *BasicScheduler) EndRun() {
	// Add an "End" event to the end of the chain
	// Then change the current event to the root of the event tree
	bs.currentEvent.AddChild(Event{Type: "End"})
	bs.currentEvent = &bs.EventRoot
}

type RunScheduler struct {
	EventRoot    tree.Tree[Event]
	currentEvent *tree.Tree[Event]

	currentSequence   *sequence.Sequence[tree.Tree[Event]]
	possibleSequences []*sequence.Sequence[tree.Tree[Event]]
}

func NewRunScheduler() *RunScheduler {
	eventTree := tree.New(Event{Type: "Start"}, EventsEquals)
	return &RunScheduler{
		EventRoot:    eventTree,
		currentEvent: &eventTree,

		currentSequence:   sequence.New(&eventTree),
		possibleSequences: make([]*sequence.Sequence[tree.Tree[Event]], 0),
	}
}

func (rs *RunScheduler) GetEvent() (*Event, error) {
	if nextEvent := rs.currentSequence.Next; nextEvent != nil {
		rs.currentSequence = nextEvent
		rs.currentEvent = nextEvent.Payload
		return &rs.currentEvent.Payload, nil
	}
	for _, child := range rs.currentEvent.Children {
		// iteratively check if each child can be the next event
		// a child can be the next event if it has some descendent leaf node that is not an "End" event
		if child.SearchLeafNodes(func(e Event) bool { return e.Type != "End" }) {
			rs.currentSequence = rs.currentSequence.InsertAfter(child)
			rs.currentEvent = child
			return &child.Payload, nil
		}
	}
	return nil, NoEventError
}

func (rs *RunScheduler) AddEvent(event Event) {
	if !rs.currentEvent.HasChild(event) {
		event := rs.currentEvent.AddChild(event)
		newSequence := rs.currentSequence.InsertAfter(event)
		rs.possibleSequences = append(rs.possibleSequences, newSequence)
	}
}

func (rs *RunScheduler) EndRun() {
	rs.currentEvent.AddChild(Event{Type: "End"})
	rs.currentEvent = &rs.EventRoot

	if len(rs.possibleSequences) > 0 {
		rs.currentSequence = rs.possibleSequences[len(rs.possibleSequences)-1]
		rs.possibleSequences = rs.possibleSequences[:len(rs.possibleSequences)-1]
	}

}
