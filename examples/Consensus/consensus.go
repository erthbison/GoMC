package main

type Value[T any] struct {
	val T
}

type HierarchicalConsensus[T any] struct {
	detectedRanks map[int]bool
	round         int
	proposal      Value[T]
	proposer      int
	proposed      bool

	DecidedSignal chan Value[T]

	// Test values used to verify algorithm
	DecidedVal  []Value[T]
	ProposedVal Value[T]

	delivered map[int]bool
	broadcast bool

	id    int
	nodes []int
	send  func(int, string, ...any)
}

func NewHierarchicalConsensus[T any](id int, nodes []int, send func(int, string, ...any)) *HierarchicalConsensus[T] {
	return &HierarchicalConsensus[T]{
		detectedRanks: make(map[int]bool),
		round:         1,
		proposal:      Value[T]{},
		proposer:      0,
		proposed:      false,
		delivered:     make(map[int]bool),
		broadcast:     false,

		DecidedSignal: make(chan Value[T], 1),
		DecidedVal:    make([]Value[T], 0),

		id:    id,
		nodes: nodes,
		send:  send,
	}
}

func (hc *HierarchicalConsensus[T]) Crash(id int) {
	hc.detectedRanks[id] = true
	for hc.delivered[hc.round] || hc.detectedRanks[hc.round] {
		hc.round++
		hc.decide()
	}
}

func (hc *HierarchicalConsensus[T]) Propose(val Value[T]) {
	hc.ProposedVal = val
	if !hc.proposed {
		hc.proposed = true
		hc.proposal = val
	}
	hc.decide()
}

func (hc *HierarchicalConsensus[T]) Decided(from int, val Value[T]) {
	if from < hc.id && from > hc.proposer {
		hc.proposed = true
		hc.proposal = val
		hc.proposer = from
		hc.decide()
	}
	hc.delivered[from] = true
	for hc.delivered[hc.round] || hc.detectedRanks[hc.round] {
		hc.round++
		hc.decide()
	}
}

func (hc *HierarchicalConsensus[T]) decide() {
	if hc.id == hc.round && hc.broadcast == false && hc.proposed {
		hc.broadcast = true
		for _, target := range hc.nodes {
			if target > hc.id {
				hc.send(int(target), "Decided", hc.id, hc.proposal)
			}
		}
		// Decide on value
		hc.DecidedSignal <- hc.proposal
		hc.DecidedVal = append(hc.DecidedVal, hc.proposal)
	}
}
