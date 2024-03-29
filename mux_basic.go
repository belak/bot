package bot

import (
	"sync"

	"github.com/sorcix/irc"
)

// BasicMux is a simple IRC event multiplexer.
// It matches the command against registered Handlers and calls the correct set.
//
// Handlers will be processed in the order in which they were added.
// Registering a handler with a "*" command will cause it to receive all events.
// Note that even though "*" will match all commands, glob matching is not used.
type BasicMux struct {
	m  map[string][]HandlerFunc
	mu *sync.Mutex
}

// This will create an initialized BasicMux with no handlers.
func NewBasicMux() *BasicMux {
	return &BasicMux{
		make(map[string][]HandlerFunc),
		&sync.Mutex{},
	}
}

// BasicMux.Event will register a Handler
func (mux *BasicMux) Event(c string, h HandlerFunc) {
	mux.mu.Lock()
	defer mux.mu.Unlock()

	mux.m[c] = append(mux.m[c], h)
}

// HandleEvent allows us to be a Handler so we can nest BasicMuxes
//
// The BasicMux simply dispatches all the Handler commands as needed
func (mux *BasicMux) HandleEvent(b *Bot, msg *irc.Message) {
	// Lock our handlers so we don't crap bricks if a
	// handler is added or removed from under our feet.
	mux.mu.Lock()
	defer mux.mu.Unlock()

	// Star means ALL THE THINGS
	// Really, this is only useful for logging
	for _, h := range mux.m["*"] {
		h(b, msg)
	}

	// Now that we've done the global handlers, we can run
	// the ones specific to this command
	for _, h := range mux.m[msg.Command] {
		h(b, msg)
	}
}
