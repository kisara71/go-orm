package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCamelToSnake(t *testing.T) {
	testCases := []struct {
		name   string
		input  string
		expect string
	}{
		{
			name:   "basic camel case",
			input:  "CamelCase",
			expect: "camel_case",
		},
		{
			name:   "ID",
			input:  "ID",
			expect: "id",
		},
		{
			name:   "with acronym",
			input:  "HTTPRequest",
			expect: "http_request",
		},
		{
			name:   "ends with number",
			input:  "UserID1",
			expect: "user_id1",
		},
		{
			name:   "mixed letters and numbers",
			input:  "simpleTest123X",
			expect: "simple_test123_x",
		},
		{
			name:   "single lower word",
			input:  "username",
			expect: "username",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res := CamelToSnake(tc.input)
			assert.Equal(t, tc.expect, res)
		})
	}
}
