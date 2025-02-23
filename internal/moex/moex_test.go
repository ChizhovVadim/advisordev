package moex

import "testing"

func TestEncodeSecurity(t *testing.T) {
	var tests = []struct {
		name string
		code string
	}{
		{
			name: "Si-6.20",
			code: "SiM0",
		},
		{
			name: "CNY-12.22",
			code: "CRZ2",
		},
	}

	for _, test := range tests {
		var code, err = EncodeSecurity(test.name)
		if err != nil {
			t.Error(test, err)
			continue
		}
		if code != test.code {
			t.Error(test, code)
			continue
		}
	}
}
