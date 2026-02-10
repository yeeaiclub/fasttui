package components

import (
	"testing"
)

func TestInputSubmitWithBackslash(t *testing.T) {
	input := NewInput()
	var submitted string

	input.onSubmit = func(value string) {
		submitted = value
	}

	input.HandleInput("h")
	input.HandleInput("e")
	input.HandleInput("l")
	input.HandleInput("l")
	input.HandleInput("o")
	input.HandleInput("\\")
	input.HandleInput("\r")

	if submitted != "hello\\" {
		t.Errorf("expected \"hello\\\\\", got %q", submitted)
	}
}

func TestInputInsertBackslash(t *testing.T) {
	input := NewInput()

	input.HandleInput("\\")
	input.HandleInput("x")

	if input.GetValue() != "\\x" {
		t.Errorf("expected \"\\\\x\", got %q", input.GetValue())
	}
}
