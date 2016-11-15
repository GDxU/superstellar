package main

//go:generate go run backend/code_generation/generate_event_dispatcher.go

import (
	"log"
	"math/rand"
	"net/http"
	"time"
)

import (
	_ "net/http/pprof"
	"superstellar/backend/communication"
	"superstellar/backend/events"
	"superstellar/backend/game"
	"superstellar/backend/monitor"
	"superstellar/backend/simulation"
	"superstellar/backend/state"
)

func main() {
	log.SetFlags(log.Lshortfile)

	rand.Seed(time.Now().UTC().UnixNano())

	eventDispatcher := events.NewEventDispatcher()
	physicsTicker := game.NewPhysicsTicker(eventDispatcher)

	monitor := monitor.NewMonitor()

	space := state.NewSpace()
	updater := simulation.NewUpdater(space, monitor, eventDispatcher)
	eventDispatcher.RegisterUserInputListener(updater)
	eventDispatcher.RegisterTimeTickListener(updater)
	eventDispatcher.RegisterUserJoinedListener(updater)
	eventDispatcher.RegisterUserLeftListener(updater)
	eventDispatcher.RegisterUserDiedListener(updater)

	srv := communication.NewServer("/superstellar", monitor, eventDispatcher)
	eventDispatcher.RegisterUserLeftListener(srv)

	sender := communication.NewSender(srv, space)
	eventDispatcher.RegisterTimeTickListener(sender)
	eventDispatcher.RegisterProjectileFiredListener(sender)
	eventDispatcher.RegisterUserLeftListener(sender)
	eventDispatcher.RegisterUserJoinedListener(sender)
	eventDispatcher.RegisterUserDiedListener(sender)

	monitor.Run()
	go srv.Listen()
	go eventDispatcher.RunEventLoop()
	go physicsTicker.Run()

	log.Fatal(http.ListenAndServe(":8080", nil))
}
