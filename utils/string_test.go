package utils

import (
	"testing"
)

func TestXmlEscape(t *testing.T) {
	input := "A \x00 terminated string."
	expected := "A \uFFFD terminated string."
	text := XmlEscape(input)
	if text != expected {
		t.Errorf("have %v, want %v", text, expected)
	}
}
