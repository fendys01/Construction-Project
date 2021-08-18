package request

import (
	"database/sql"
	"encoding/json"
	"panorama/services/api/model"
	"strconv"
)

type CitCallLogSMSOrOTPReq struct {
	TrxID        string `json:"trxid"`
	Result       string `json:"result"`
	Description  string `json:"description"`
	ReportedDate string `json:"reportedDate"`
	Currency     string `json:"currency"`
	Price        string `json:"price"`
}

type CitCallLogCallReq struct {
	TrxID      string `json:"trxid"`
	Result     string `json:"result"`
	RC         string `json:"rc"`
	Msisdn     string `json:"msisdn"`
	Via        string `json:"via"`
	Token      string `json:"token"`
	DialCode   string `json:"dial_code"`
	DialStatus string `json:"dial_status"`
	CallStatus string `json:"call_status"`
	Price      string `json:"price"`
}

func CitCallLog(trxID, callType, price string, reqLog interface{}) (model.CallLogEnt, error) {
	cl := model.CallLogEnt{}

	cl.TrxID = trxID
	cl.Provider = model.PROVIDER_CITCALL
	cl.CallType = callType

	// Assign bill_price
	billPrice, err := strconv.Atoi(price)
	if err != nil {
		return cl, err
	}
	cl.BillPrice = sql.NullInt64{Int64: int64(billPrice), Valid: true}

	// Assign payloads
	encode, _ := json.Marshal(reqLog)
	resArray := make(map[string]interface{})
	err = json.Unmarshal(encode, &resArray)
	if err != nil {
		return cl, err
	}
	cl.Payloads = resArray

	return cl, nil
}
