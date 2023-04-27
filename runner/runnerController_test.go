package runner

import (
	"gomc/event"
	"strconv"
	"testing"
)

func TestCommandsIncorrectId(t *testing.T) {
	for i, test := range commandTest {
		ec := NewEventController[MockNode, int](1000)

		nodes := make(map[int]*MockNode)
		for _, id := range test.nodeIds {
			nodes[id] = &MockNode{Id: id}
		}
		ec.MainLoop(nodes, eventBuffer, crashFunc, getState)
		var err error
		for _, cmd := range test.commands {
			switch cmd.cmd {
			case "stop":
				ec.Stop()
			case "pause":
				err = ec.Pause(cmd.id)
			case "resume":
				err = ec.Resume(cmd.id)
			case "crash":
				err = ec.CrashNode(cmd.id)
			}
		}
		if err == nil {
			t.Errorf("Test %v: Expected to receive an error", i)
		}
		ec.Stop()
	}
}

func TestCrashSubscribe(t *testing.T) {
	for i, test := range crashNodeTest {
		ec := NewEventController[MockNode, int](1000)
		nodes := make(map[int]*MockNode)
		for _, id := range test.nodeIds {
			nodes[id] = &MockNode{Id: id}
		}

		ec.MainLoop(nodes, eventBuffer, crashFunc, getState)

		nodeId := -1
		newStatus := true
		crashSubscribe := func(id int, status bool) {
			nodeId = id
			newStatus = status
		}

		ec.CrashSubscribe(0, crashSubscribe)

		err := ec.CrashNode(test.id)
		isErr := (err != nil)
		if isErr != test.expectErr {
			if isErr {
				t.Errorf("Test %v: Expected no error got: %v", i, err)
			} else {
				t.Errorf("Test %v: Expected to receive an error", i)
			}
		}

		if !test.expectErr {
			if newStatus != false {
				t.Errorf("Test %v: EventController: Expected callback to be called and status to be updated. Got: %v", i, newStatus)
			}
			if nodeId != test.id {
				t.Errorf("Test %v: EventController: Expected callback to be called and nodeId to be updated. Got: %v", i, nodeId)
			}
		} else {
			if newStatus != true {
				t.Errorf("Test %v: EventController: Expected callback not to be called. Status have been updated. Got: %v", i, newStatus)
			}
			if nodeId != -1 {
				t.Errorf("Test %v: EventController: Expected callback not to be called. NodeId has been updated. Got: %v", i, nodeId)
			}
		}

	}
}

func TestAddEvent(t *testing.T) {
	for _, test := range addEventTest {
		ec := NewEventController[MockNode, int](1000)
		nodes := make(map[int]*MockNode)
		for _, id := range test.nodeIds {
			nodes[id] = &MockNode{Id: id}
		}

		ec.MainLoop(nodes, eventBuffer, crashFunc, getState)
		recSub := ec.Subscribe()
		for j, target := range test.eventTargets {
			ec.AddEvent(MockEvent{
				id:     event.EventId(strconv.Itoa(j)),
				target: target,
			})
		}

		ec.Stop()

		numEvt := 0
		for j := 0; j < len(test.eventTargets)*2; j++ {
			rec := <-recSub
			if rec, ok := rec.(ExecutionRecord); ok {
				expectedEvt := test.eventTargets[numEvt]
				evtId, err := strconv.Atoi(string(rec.Evt.Id()))
				if err != nil {
					panic(err)
				}
				if expectedEvt != evtId {
					t.Errorf("Unexpected event. Expected %v. Got: %v", expectedEvt, evtId)
				}
				numEvt++
			}
		}
	}
}

func TestMultipleClose(t *testing.T) {
	ec := NewEventController[MockNode, int](1000)
	nodes := make(map[int]*MockNode)
	for id := 0; id < 3; id++ {
		nodes[id] = &MockNode{Id: id}
	}
	ec.MainLoop(nodes, eventBuffer, crashFunc, getState)
	ec.Stop()

	ec.Stop()
}

type command struct {
	cmd string
	id  int
}

var commandTest = []struct {
	nodeIds  []int
	commands []command
}{
	{
		[]int{0, 1, 2},
		[]command{{"pause", 10}},
	},
	{
		[]int{0, 1, 2},
		[]command{{"stop", -1}, {"pause", 0}},
	},
	{
		[]int{0, 1, 2},
		[]command{{"resume", 10}},
	},
	{
		[]int{0, 1, 2},
		[]command{{"stop", -1}, {"resume", 0}},
	},
	{
		[]int{0, 1, 2},
		[]command{{"crash", 10}},
	},
	{
		[]int{0, 1, 2},
		[]command{{"stop", -1}, {"crash", 1}},
	},
}

var crashNodeTest = []struct {
	nodeIds []int
	id      int

	expectErr bool
}{
	{
		[]int{0, 1, 2},
		1,

		false,
	},
	{
		[]int{0, 1, 2},
		10,

		true,
	},
}

var addEventTest = []struct {
	nodeIds      []int
	eventTargets []int

	expectedEvents []int
}{
	{
		[]int{0, 1, 2},
		[]int{0, 1},

		[]int{0, 1},
	},
	{
		[]int{0, 1, 2},
		[]int{0, 10, 1},
		[]int{0, 2},
	},
}
