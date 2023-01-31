package tester

// import (
// 	"experimentation/sequence"
// 	"experimentation/tree"
// )

// type Checker[S comparable] interface {
// 	Check(*tree.Tree[[]S]) (bool, sequence.Sequence[S])
// }

// type PredicateChecker[S comparable] struct {
// 	// A slice of predicates that returns true if the state breaks the predicate and false of it does not
// 	predicates []func(S) bool
// }

// func (pc *PredicateChecker[S]) Check(stateTree *tree.Tree[[]S]) (bool, sequence.Sequence[S]) {

// }

// func (pc *PredicateChecker[S]) checkState(state *S) bool {
// 	for _, pred := range pc.predicates {
// 		if !pred(*state) {
// 			return false
// 		}
// 	}
// 	return true
// }
