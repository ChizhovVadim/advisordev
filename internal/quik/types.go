package quik

import "encoding/json"

type RequestJson struct {
	Id          int64       `json:"id"`
	Command     string      `json:"cmd"`
	CreatedTime int64       `json:"t"`
	Data        interface{} `json:"data"`
}

type ResponseJson struct {
	Id          int64            `json:"id"`
	Command     string           `json:"cmd"`
	CreatedTime float64          `json:"t"`
	Data        *json.RawMessage `json:"data"`
	LuaError    string           `json:"lua_error"`
}

type CallbackJson struct {
	Command     string           `json:"cmd"`
	CreatedTime float64          `json:"t"`
	Data        *json.RawMessage `json:"data"`
	LuaError    string           `json:"lua_error"`
}

type CandleInterval int

const (
	CandleIntervalM5 CandleInterval = 5
)

type Candle struct {
	Low       float64        `json:"low"`
	Close     float64        `json:"close"`
	High      float64        `json:"high"`
	Open      float64        `json:"open"`
	Volume    float64        `json:"volume"`
	Datetime  QuikDateTime   `json:"datetime"`
	SecCode   string         `json:"sec"`
	ClassCode string         `json:"class"`
	Interval  CandleInterval `json:"interval"`
}

type QuikDateTime struct {
	Ms    int `json:"ms"`
	Sec   int `json:"sec"`
	Min   int `json:"min"`
	Hour  int `json:"hour"`
	Day   int `json:"day"`
	Month int `json:"month"`
	Year  int `json:"year"`
}

type TradeEventData struct {
	TradeNum int64        `json:"trade_num"`
	Account  string       `json:"account"`
	Price    float64      `json:"price"`
	Quantity float64      `json:"qty"`
	SecCode  string       `json:"sec_code"`
	DateTime QuikDateTime `json:"datetime"`
	TransID  int64        `json:"trans_id"`
}

const (
	EventNameOnConnected    = "OnConnected"
	EventNameOnDisconnected = "OnDisconnected"
	EventNameOnTrade        = "OnTrade"
	EventNameNewCandle      = "NewCandle"
)
