package cmd

import (
	"reflect"
	"testing"
)

func TestExecCmdLines(t *testing.T) {
	result, err := execCmdLines("echo a; echo b")
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	expected := []string{"a", "b"}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("got %v while expected %v", result, expected)
	}
}
