package ftx

type walletResponse struct {
	Result map[string][]struct {
		Coin  string
		Total float64
	}
}
