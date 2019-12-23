package main

import "testing"

import "github.com/stretchr/testify/assert"

func TestMergeRequests(t *testing.T) {
	p := DispatchRequest{}
	s := DispatchRequest{}
	p["item0"] = "val0"
	p["item1"] = "val1"
	s["item0"] = "val2"

	expected := map[string]string{
		"item0": "val0",
		"item1": "val1",
	}

	m := mergeRequests(p, s)

	assert.EqualValues(t, expected, m)

}
