package quik

import (
	"fmt"
	"strconv"
)

func AsInt(a any) (int, error) {
	switch v := a.(type) {
	case float64:
		return int(v), nil
	case string:
		var f, err = strconv.Atoi(v)
		if err != nil {
			return 0, fmt.Errorf("wrong value type %v", v)
		}
		return f, nil
	default:
		return 0, fmt.Errorf("wrong value type %v", v)
	}
}

func AsFloat64(a any) (float64, error) {
	switch v := a.(type) {
	case float64:
		return v, nil
	case string:
		var f, err = strconv.ParseFloat(v, 64)
		if err != nil {
			return 0, fmt.Errorf("wrong value type %v", v)
		}
		return f, nil
	default:
		return 0, fmt.Errorf("wrong value type %v", v)
	}
}
