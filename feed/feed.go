package feed

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) Add(chanF func() chan float64, processF func(float64)) {
	go func() {
		ch := chanF()
		for {
			processF(<-ch)
		}
	}()
}

func (h *Handler) AddArray(chanF func() chan [2]float64, processF func([2]float64)) {
	go func() {
		ch := chanF()
		for {
			processF(<-ch)
		}
	}()
}
