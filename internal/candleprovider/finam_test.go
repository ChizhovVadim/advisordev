package candleprovider

import (
	"net/http"
	"testing"
	"time"
)

func TestFinamHistoryCandles(t *testing.T) {
	var client = &http.Client{Timeout: 25 * time.Second}
	const finamSecurityCode = 909980 //"Si-9.21"
	var end = time.Now()
	var start = end.AddDate(0, 0, -3)
	var candles, err = finamGet(client, finamSecurityCode, finamPeriodMinutes5, start, end)
	if err != nil {
		t.Error(err)
		return
	}
	if len(candles) == 0 {
		t.Error("empty candles")
		return
	}
	t.Log(candles)
}
