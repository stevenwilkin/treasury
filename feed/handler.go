package feed

import (
	"reflect"
	"sync"
	"time"
)

type Handler struct {
	feeds Status
	m     sync.Mutex
}

type FeedStatus struct {
	Active     bool
	LastUpdate time.Time
}
type Status map[Feed]*FeedStatus

func NewHandler() *Handler {
	return &Handler{
		feeds: Status{}}
}

func (h *Handler) setActive(f Feed, active bool) {
	h.m.Lock()
	h.feeds[f].Active = active
	h.m.Unlock()
}

func (h *Handler) setLastUpdate(f Feed) {
	h.m.Lock()
	h.feeds[f].LastUpdate = time.Now()
	h.m.Unlock()
}

func (h *Handler) Add(f Feed, inputF interface{}, outputF interface{}) {
	h.m.Lock()
	defer h.m.Unlock()

	h.feeds[f] = &FeedStatus{Active: true}

	go func() {
		ch := reflect.ValueOf(inputF).Call([]reflect.Value{})[0]
		for {
			item, ok := ch.Recv()
			if !ok {
				h.setActive(f, false)
				return
			}

			h.setLastUpdate(f)
			reflect.ValueOf(outputF).Call([]reflect.Value{item})
		}
	}()
}

func (h *Handler) Status() Status {
	return h.feeds
}
