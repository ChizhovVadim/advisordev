package quik

import (
	"fmt"
)

type Nullable struct {
	valid bool
}

func (n *Nullable) Valid() bool {
	return n.valid
}

func (n *Nullable) SetValid(v bool) {
	n.valid = v
}

func (quik *QuikService) IsConnected() (bool, error) {
	var resp int
	var err = quik.ExecuteQuery(
		"isConnected",
		"",
		&resp)
	return resp == 1, err
}

func (quik *QuikService) MessageInfo(msg string) error {
	return quik.ExecuteQuery("message", msg, nil)
}

type GetPortfolioInfoExRequest struct {
	FirmId     string
	ClientCode string
	LimitKind  int
}

type GetPortfolioInfoExResponse struct {
	Nullable
	StartLimitOpenPos string `json:"start_limit_open_pos"`
}

func (quik *QuikService) GetPortfolioInfoEx(
	req GetPortfolioInfoExRequest) (GetPortfolioInfoExResponse, error) {
	var resp GetPortfolioInfoExResponse
	var err = quik.ExecuteQuery(
		"getPortfolioInfoEx",
		fmt.Sprintf("%v|%v|%v", req.FirmId, req.ClientCode, req.LimitKind),
		&resp)
	return resp, err
}

type GetFuturesHoldingRequest struct {
	FirmId  string
	AccId   string
	SecCode string
	PosType int
}

type GetFuturesHoldingResponse struct {
	Nullable
	TotalNet float64 `json:"totalnet"`
}

func (quik *QuikService) GetFuturesHolding(
	req GetFuturesHoldingRequest) (GetFuturesHoldingResponse, error) {
	var resp GetFuturesHoldingResponse
	var err = quik.ExecuteQuery(
		"getFuturesHolding",
		fmt.Sprintf("%v|%v|%v|%v", req.FirmId, req.AccId, req.SecCode, req.PosType),
		&resp)
	return resp, err
}

type Transaction struct {
	TRANS_ID    string
	ACTION      string
	ACCOUNT     string
	CLASSCODE   string
	SECCODE     string
	QUANTITY    string
	OPERATION   string
	PRICE       string
	CLIENT_CODE string
}

func (quik *QuikService) SendTransaction(
	req Transaction) error {
	quik.transId += 1
	req.TRANS_ID = fmt.Sprintf("%v", quik.transId)
	req.CLIENT_CODE = req.TRANS_ID
	var resp bool
	return quik.ExecuteQuery(
		"sendTransaction",
		req,
		&resp)
}

func (quik *QuikService) GetLastCandles(
	classCode, securityCode string, interval CandleInterval, count int) ([]Candle, error) {
	var resp []Candle
	var err = quik.ExecuteQuery(
		"get_candles_from_data_source",
		fmt.Sprintf("%v|%v|%v|%v", classCode, securityCode, interval, count),
		&resp)
	return resp, err
}

func (quik *QuikService) SubscribeCandles(
	classCode, securityCode string, interval CandleInterval) error {
	var resp string
	return quik.ExecuteQuery(
		"subscribe_to_candles",
		fmt.Sprintf("%v|%v|%v", classCode, securityCode, interval),
		&resp)
}

// scale - Количество значащих цифр после запятой
// mat_date - Дата погашения (число YYYYMMDD)
// lot_size - Размер лота
// min_price_step - Минимальный шаг цены
func GetSecurityInfo(
	quik *QuikService,
	classCode string,
	secCode string,
) (map[string]any, error) {
	var resp, err = quik.ExecuteQueryDynamic("getSecurityInfo",
		fmt.Sprintf("%v|%v", classCode, secCode))
	var res, _ = resp.(map[string]any)
	return res, err
}

const (
	//Цена последней сделки
	ParamNameLAST = "LAST"
	/// Лучшая цена предложения
	ParamNameOFFER = "OFFER"
	/// Лучшая цена спроса
	ParamNameBID = "BID"
	/// Максимально возможная цена
	ParamNamePRICEMAX = "PRICEMAX"
	/// Минимально возможная цена
	ParamNamePRICEMIN = "PRICEMIN"
	/// Гарантийное обеспечение покуптеля
	ParamNameBUYDEPO = "BUYDEPO"
	/// Гарантийное обеспечение продавца
	ParamNameSELLDEPO = "SELLDEPO"
)

// param_value
func GetParamEx(
	quik *QuikService,
	classCode string,
	secCode string,
	paramName string,
) (map[string]any, error) {
	var resp, err = quik.ExecuteQueryDynamic("getParamEx",
		fmt.Sprintf("%v|%v|%v", classCode, secCode, paramName))
	var res, _ = resp.(map[string]any)
	return res, err
}
