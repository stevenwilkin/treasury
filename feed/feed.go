package feed

import (
	"reflect"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) Add(inputF interface{}, outputF interface{}) {
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
