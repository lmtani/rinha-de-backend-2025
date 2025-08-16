package entity

type Payment struct {
	CorrelationId string `json:"correlationId"`
	Amount        string `json:"amount"`
}

type PaymentsChannelStats struct {
	TotalRequests int     `json:"totalRequests"`
	TotalAmount   float64 `json:"totalAmount"`
}

type PaymentsSummary struct {
	Default  PaymentsChannelStats `json:"default"`
	Fallback PaymentsChannelStats `json:"fallback"`
}
