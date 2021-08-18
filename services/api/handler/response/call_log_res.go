package response

import (
	"panorama/services/api/model"
	"time"
)

// CallLogResponse ...
type CallLogResponse struct {
	ID          int32                  `json:"id"`
	TrxID       string                 `json:"trx_id"`
	Provider    string                 `json:"provider"`
	CallType    string                 `json:"call_type"`
	BillPrice   int64                  `json:"bill_price"`
	Payloads    map[string]interface{} `json:"payloads"`
	CreatedDate time.Time              `json:"created_date"`
}

// Transform from call logs
func (c CallLogResponse) Transform(i model.CallLogEnt) CallLogResponse {
	c.ID = i.ID
	c.TrxID = i.TrxID
	c.Provider = i.Provider
	c.CallType = i.CallType
	c.BillPrice = i.BillPrice.Int64
	c.Payloads = i.Payloads
	c.CreatedDate = i.CreatedDate

	return c
}
