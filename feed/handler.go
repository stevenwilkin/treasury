package feed

import (
	"math"
	"reflect"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

var (
	delayBase  = 2.0
	maxRetries = 6
)

type Handler struct {
	feeds Status
	m     sync.Mutex
}

type FeedStatus struct {
	inputF     interface{}
	outputF    interface{}
	Active     bool
	LastUpdate time.Time
	Errors     int
}
type Status map[Feed]*FeedStatus

func NewHandler() *Handler {
	return &Handler{
		feeds: Status{}}
}

func (h *Handler) feedStatus(f Feed) FeedStatus {
	h.m.Lock()
	defer h.m.Unlock()

	return *h.feeds[f]
}

func (h *Handler) startFeed(f Feed) reflect.Value {
	fn := h.feedStatus(f).inputF
	return reflect.ValueOf(fn).Call([]reflect.Value{})[0]
}

func (h *Handler) setFailed(f Feed) {
	h.m.Lock()
	h.feeds[f].Active = false
	h.feeds[f].Errors += 1
	h.m.Unlock()
}

func (h *Handler) processFeed(f Feed, item reflect.Value) {
	h.m.Lock()
	defer h.m.Unlock()

	h.feeds[f].LastUpdate = time.Now()
	h.feeds[f].Active = true
	h.feeds[f].Errors = 0

	fn := h.feeds[f].outputF
	reflect.ValueOf(fn).Call([]reflect.Value{item})
}

func (h *Handler) canRestart(f Feed) bool {
	return h.feedStatus(f).Errors <= maxRetries
}

func (h *Handler) exponentialBackoff(f Feed) {
	delaySeconds := math.Pow(delayBase, float64(h.feedStatus(f).Errors))
	delay := time.Second * time.Duration(delaySeconds)
	log.WithFields(log.Fields{
		"feed":  f,
		"delay": delay,
	}).Warn("Backing off feed")
	time.Sleep(delay)
}

func (h *Handler) Add(f Feed, inputF interface{}, outputF interface{}) {
	h.m.Lock()
	defer h.m.Unlock()

	h.feeds[f] = &FeedStatus{
		inputF:  inputF,
		outputF: outputF,
		Active:  true}

	go func() {
		ch := h.startFeed(f)
		for {
			item, ok := ch.Recv()
			if !ok {
				h.setFailed(f)
				if h.canRestart(f) {
					h.exponentialBackoff(f)
					ch = h.startFeed(f)
				} else {
					log.WithField("feed", f).Error("Feed failed")
					return
				}
			} else {
				h.processFeed(f, item)
			}
		}
	}()
}

func (h *Handler) Status() map[Feed]FeedStatus {
	h.m.Lock()
	defer h.m.Unlock()

	result := map[Feed]FeedStatus{}
	for f, s := range h.feeds {
		result[f] = *s
	}

	return result
}
