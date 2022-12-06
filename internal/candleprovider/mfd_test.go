package candleprovider

import (
	"net/http"
	"testing"
	"time"
)

func TestMfdHistoryCandles(t *testing.T) {
	var client = &http.Client{Timeout: 25 * time.Second}
	const mfdSecurityCode = 29418 //"Si-9.21"
	var end = time.Now()
	var start = end.AddDate(0, 0, -3)
	var candles, err = mfdGet(client, mfdSecurityCode, mfdPeriodMinutes5, start, end)
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
