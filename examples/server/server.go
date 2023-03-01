package server

import (
	"encoding/gob"
	"errors"
	"fmt"
	"log"
	"net"
	"reflect"
)

func RegisterTypes(types ...any) {
	for _, t := range types {
		gob.Register(t)
	}
}

type Message struct {
	Type   string
	Module string
	Params []any
}

type Indication struct {
	Type   string
	Params []any
}

type Server struct {
	modules map[string]any

	msg        chan Message
	indication chan Indication
	connection map[int]net.Conn

	// Modules that have subscribed to indications from the server
	subscriptions map[string][]string

	pause    chan bool
	resume   chan bool
	isPaused bool
}

func NewServer() *Server {
	return &Server{
		modules:       make(map[string]any),
		indication:    make(chan Indication),
		msg:           make(chan Message),
		connection:    make(map[int]net.Conn),
		subscriptions: make(map[string][]string),

		pause:    make(chan bool),
		resume:   make(chan bool),
		isPaused: false,
	}
}

func (s *Server) RegisterModule(name string, module any) error {
	if _, ok := s.modules[name]; ok {
		return fmt.Errorf("A module with the name %q already exists", name)
	}
	s.modules[name] = module
	return nil
}

func (s *Server) Start(lis net.Listener) {
	go func() {
		for {
			conn, err := lis.Accept()
			if err != nil {
				log.Printf("Received error while listening: %v", err)
			}
			go s.handleConnection(conn)
		}
	}()
	s.startMainLoop()
}

func (s *Server) startMainLoop() {
	go func() {
		for {
			// Pause the main loop until a resume indication is received
			select {
			case <-s.pause:
				s.isPaused = true
				<-s.resume
				s.isPaused = false
			default:
			}

			select {
			case msg := <-s.msg:
				err := s.handleMessages(msg)
				if err != nil {
					log.Print(err)
				}
			case indication := <-s.indication:
				err := s.handleIndication(indication)
				if err != nil {
					log.Print(err)
				}
			}
		}
	}()
}

func (s *Server) DialNodes(addresses map[int]string, dial func(string) (net.Conn, error)) error {
	for id, addr := range addresses {
		conn, err := dial(addr)
		if err != nil {
			return fmt.Errorf("Received error while dialing node: %v", err)
		}
		s.connection[id] = conn
	}
	return nil
}

func (s *Server) handleConnection(conn net.Conn) error {
	dec := gob.NewDecoder(conn)
	for {
		var msg Message
		err := dec.Decode(&msg)
		if errors.Is(err, net.ErrClosed) {
			break
		} else if err != nil {
			return fmt.Errorf("Error while deserializing the message: %v", err)
		}
		s.msg <- msg
	}
	return nil
}

func (s *Server) handleMessages(msg Message) error {
	module := s.modules[msg.Module]
	valueParams := make([]reflect.Value, len(msg.Params))
	for i, val := range msg.Params {
		valueParams[i] = reflect.ValueOf(val)
	}
	method := reflect.ValueOf(module).MethodByName(msg.Type)
	method.Call(valueParams)
	return nil
}

func (s *Server) handleIndication(ind Indication) error {
	valueParams := make([]reflect.Value, len(ind.Params))
	for i, val := range ind.Params {
		valueParams[i] = reflect.ValueOf(val)
	}
	modules, ok := s.subscriptions[ind.Type]
	if !ok {
		return fmt.Errorf("No modules subscribed to indication %q", ind.Type)
	}
	for _, moduleName := range modules {
		module, ok := s.modules[moduleName]
		if !ok {
			return fmt.Errorf("Unable to find the module: %q", moduleName)
		}
		method := reflect.ValueOf(module).MethodByName(ind.Type)
		method.Call(valueParams)
	}
	return nil
}

func (s *Server) SendFactory(module string) func(target int, msgType string, params ...interface{}) {
	return func(target int, msgType string, params ...interface{}) {
		conn, ok := s.connection[target]
		if !ok {
			log.Panicf("No connection to target: %v", target)
		}
		enc := gob.NewEncoder(conn)

		err := enc.Encode(&Message{
			Type:   msgType,
			Params: params,
			Module: module,
		})
		if err != nil {
			log.Panicf("Error while serializing the message: %v", err)
		}
	}
}

func (s *Server) Indicate(indicationType string, params ...interface{}) {
	s.indication <- Indication{
		Type:   indicationType,
		Params: params,
	}
}

func (s *Server) Subscribe(indication string, module string) {
	modules, ok := s.subscriptions[indication]
	if !ok {
		modules = make([]string, 0)
	}
	modules = append(modules, module)
	s.subscriptions[indication] = modules
}

func (s *Server) Pause() error {
	if s.isPaused {
		return fmt.Errorf("Node is already paused")
	}
	s.pause <- true
	return nil
}

func (s *Server) Resume() error {
	if !s.isPaused {
		return fmt.Errorf("Node is not paused")
	}
	s.resume <- true
	return nil
}
