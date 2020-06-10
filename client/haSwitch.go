package client

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type State struct {
	state bool
	mux   sync.Mutex
}

func (s *State) Get() bool {
	s.mux.Lock()
	defer s.mux.Unlock()
	return s.state
}

func (s *State) Set(val bool) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.state = val
}

type HASwitch struct {
	state      *State
	clientPort string
	client     *WHControlClient
	bootGrace  *State
	bootTime   int
}

func NewHASwitch(clientPort string, client *WHControlClient, bootTime int) *HASwitch {
	return &HASwitch{&State{state: false}, clientPort, client, &State{state: false}, bootTime}
}

func (has *HASwitch) Start() error {
	http.HandleFunc("/", has.serve)
	go has.client.Start(has.state, has.bootGrace)
	return http.ListenAndServe(":"+has.clientPort, nil)
}

func (has *HASwitch) serve(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST", "PUT":
		bytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			break
		}
		newState, err := strconv.ParseBool(string(bytes))
		if err != nil {
			break
		}
		currentState := has.state.Get()
		if newState != currentState {
			if newState {
				if err := has.client.Wake(); err != nil {
					log.Println("Failed to wake", err)
				}
				has.state.Set(newState)
			} else {
				if err := has.client.Hibernate(); err != nil {
					log.Println("Failed to hibernate", err)
				}
				has.bootGrace.Set(true)
				go func() {
					time.Sleep(time.Duration(has.bootTime) * time.Second)
					has.bootGrace.Set(false)
				}()
				has.state.Set(newState)
			}
		}
	}
	fmt.Fprintf(w, "%t\n", has.state.Get())
}
