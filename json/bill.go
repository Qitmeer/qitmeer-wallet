package json

type PaymentResult struct {
	TxID      string `json:"tx_id"`
	Variation int64  `json:"variation"`
}
type BillResult []PaymentResult

type PagedBillResult struct {
	Total    int32      `json:"total"`
	PageNo   int32      `json:"page_no"`
	PageSize int32      `json:"page_size"`
	Bill     BillResult `json:"bill,omitempty"`
}
