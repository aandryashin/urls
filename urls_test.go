package main

import (
	"testing"

	. "github.com/aandryashin/matchers"
)

func TestFunc(t *testing.T) {
	AssertThat(t, true, Is{true})
}
