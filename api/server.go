package api

import (
	"context"
)

type Server struct {
	Broker *Broker
	// add DB, logger, etc.
}

type InsightEvent struct {
	Name   string
	UserID string      // user id this is for
	Data   interface{} // JSON-serializable payload
}

type Subscriber struct {
	ch   chan InsightEvent
	done <-chan struct{}
}

type cmd struct {
	kind string
	user string
	sub  *Subscriber
	ev   InsightEvent
}

// TODO use mutex to prevent race conditions / panics
type Broker struct {
	cmds chan cmd
}

// NewBroker starts a single goroutine that owns the subscriber maps (no mutex needed).
func NewBroker(buffer int) *Broker {
	b := &Broker{cmds: make(chan cmd, 1024)}
	subs := map[string]map[*Subscriber]struct{}{}

	go func() {
		for c := range b.cmds {
			switch c.kind {
			case "sub":
				if subs[c.user] == nil {
					subs[c.user] = map[*Subscriber]struct{}{}
				}
				subs[c.user][c.sub] = struct{}{}
			case "unsub":
				if m := subs[c.user]; m != nil {
					delete(m, c.sub)
					if len(m) == 0 {
						delete(subs, c.user)
					}
					close(c.sub.ch) // ok: only broker closes
				}
			case "pub":
				for s := range subs[c.ev.UserID] {
					select {
					case s.ch <- c.ev:
					default: // drop if slow (prototype-friendly)
					}
				}
			}
		}
	}()

	return b
}

func (b *Broker) Subscribe(ctx context.Context, user string, buf int) *Subscriber {
	sub := &Subscriber{ch: make(chan InsightEvent, buf), done: ctx.Done()}
	b.cmds <- cmd{kind: "sub", user: user, sub: sub}
	go func() {
		<-ctx.Done()
		b.cmds <- cmd{kind: "unsub", user: user, sub: sub}
	}()
	return sub
}

func (b *Broker) Publish(ev InsightEvent) { b.cmds <- cmd{kind: "pub", ev: ev} }
