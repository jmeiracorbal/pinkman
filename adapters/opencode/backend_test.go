package opencode

import (
	"reflect"
	"testing"
)

func TestParseModels(t *testing.T) {
	got := ParseModels("anthropic/claude\nopenai/gpt\nanthropic/claude\n\n")
	want := []string{"anthropic/claude", "openai/gpt"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ParseModels() = %v, want %v", got, want)
	}
}
