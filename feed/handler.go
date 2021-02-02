package feed

import (
	"reflect"
)

type Handler struct {
	feeds Status
}

type FeedStatus bool
type Status map[Feed]FeedStatus

func NewHandler() *Handler {
	return &Handler{
		feeds: Status{}}
}

func (h *Handler) Add(f Feed, inputF interface{}, outputF interface{}) {
	h.feeds[f] = true

	go func() {
		ch := reflect.ValueOf(inputF).Call([]reflect.Value{})[0]
		for {
			item, ok := ch.Recv()
			if !ok {
				return
			}
			reflect.ValueOf(outputF).Call([]reflect.Value{item})
		}
	}()
}

func (h *Handler) Status() Status {
	return h.feeds
}
