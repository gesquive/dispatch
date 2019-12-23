package main

import "testing"

import "net/http"
import "github.com/stretchr/testify/assert"

func TestGetHeaders(t *testing.T) {
	headers := http.Header{}
	headers.Add("x-dispatch-auth-token", "123-456")
	headers.Add("x-dispatch-subject", "Test!")
	headers.Add("x-dispatch-special", "value")
	headers.Add("from", "no-one")

	expected := map[string]string{
		"auth-token": "123-456",
		"subject":    "Test!",
		"special":    "value",
	}

	result := getHeaderValues(headers)

	assert.EqualValues(t, expected, result)
}
