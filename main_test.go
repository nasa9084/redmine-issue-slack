package main

import "testing"

func TestExtractTicketID(t *testing.T) {
	candidates := []struct {
		in     string
		expect int
	}{
		{"", -1},
		{"foo", -1},
		{"#123", 123},
		{"foo#123", 123},
		{"#123bar", 123},
		{"foo#123bar", 123},
		{"foo#123bar#456baz", 123},
		{"ほげ #123", 123},
	}
	for _, c := range candidates {
		if out := extractTicketID(c.in); out != c.expect {
			t.Errorf("input(%s): %d != %d", c.in, out, c.expect)
		}
	}
}
