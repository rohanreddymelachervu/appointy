package main

import (
	"testing"
)

func TestHelloWorld(t *testing.T) {
	actual := main.newConnectToMongoDB()
	if actual != "hello world" {
		t.Errorf("expected 'hello world', got '%s'", actual)
	}
}
