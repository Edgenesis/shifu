package main

import (
	"context"
	"fmt"
	"github.com/edgenesis/shifu/pkg/deviceshifu/mockdevice/mockdevice"
	"github.com/edgenesis/shifu/pkg/logger"
	"github.com/looplab/fsm"
	"math/rand"
	"net/http"
)

const (
	RED    string = "RED"
	YELLOW string = "YELLOW"
	GREEN  string = "GREEN"
)

const (
	STOP    string = "STOP"
	CAUTION string = "CAUTION"
	PROCEED string = "PROCEED"
)

type TrafficLight struct {
	FSM *fsm.FSM
}

func NewTrafficLight(color string) *TrafficLight {
	tl := &TrafficLight{}

	tl.FSM = fsm.NewFSM(
		color,
		fsm.Events{
			{Name: STOP, Src: []string{YELLOW}, Dst: RED},
			{Name: CAUTION, Src: []string{GREEN}, Dst: YELLOW},
			{Name: PROCEED, Src: []string{RED}, Dst: GREEN},
		},
		fsm.Callbacks{},
	)

	return tl
}

var trafficLight *TrafficLight

func main() {
	availableFuncs := []string{
		"stop",
		"caution",
		"proceed",
		"get_color",
		"get_status",
	}
	trafficLight = NewTrafficLight(RED)
	mockdevice.StartMockDevice(availableFuncs, instructionHandler)
}

func instructionHandler(functionName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Infof("Handling: %v", functionName)
		switch functionName {
		case "stop":
			err := trafficLight.FSM.Event(context.Background(), STOP)

			if err != nil {
				logger.Warnf("Disable transition from %v to %v, must be %v", trafficLight.FSM.Current(), RED, YELLOW)
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprintf(w, "Disable transition from %v to %v, must be %v", trafficLight.FSM.Current(), RED, YELLOW)
				return
			}
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "Transition from %v to %v", YELLOW, trafficLight.FSM.Current())
		case "caution":
			err := trafficLight.FSM.Event(context.Background(), CAUTION)
			if err != nil {
				logger.Warnf("Disable transition from %v to %v, must be %v", trafficLight.FSM.Current(), YELLOW, GREEN)
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprintf(w, "Disable transition from %v to %v, must be %v", trafficLight.FSM.Current(), YELLOW, GREEN)
				return
			}
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "Transition from %v to %v", GREEN, trafficLight.FSM.Current())
		case "proceed":
			err := trafficLight.FSM.Event(context.Background(), PROCEED)
			if err != nil {
				logger.Warnf("Disable transition from %v to %v, must be %v", trafficLight.FSM.Current(), GREEN, RED)
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprintf(w, "Disable transition from %v to %v, must be %v", trafficLight.FSM.Current(), GREEN, RED)
				return
			}
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "Transition from %v to %v", RED, trafficLight.FSM.Current())
		case "get_color":
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "traffic light current state: %v", trafficLight.FSM.Current())
		case "get_status":
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, mockdevice.StatusSetList[(rand.Intn(len(mockdevice.StatusSetList)))])
		}
	}
}
