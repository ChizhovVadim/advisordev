package quik

import "testing"

func TestRound(t *testing.T) {
	var tests = []struct {
		val       float64
		precision int
		expected  string
	}{
		{
			val:       8.123456,
			precision: 0,
			expected:  "8",
		},
		{
			val:       8.123456,
			precision: 1,
			expected:  "8.1",
		},
		{
			val:       8.123456,
			precision: 2,
			expected:  "8.12",
		},
		{
			val:       8.123456,
			precision: 3,
			expected:  "8.123",
		},
	}
	for _, test := range tests {
		var y = formatPrice(0, test.precision, test.val)
		if y != test.expected {
			t.Error(test, y)
			continue
		}
	}
}
