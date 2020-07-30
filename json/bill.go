package json

type BillResult struct {
	TxID        string `json:"tx_id"`
	Amount      int64  `json:"amount"`
}
type BillsResult []BillResult

type PagedBillsResult struct {
	Total    int32       `json:"total"`
	PageNo   int32       `json:"page_no"`
	PageSize int32       `json:"page_size"`
	Bills    BillsResult `json:"bills"`
}
