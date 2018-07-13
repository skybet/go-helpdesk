package wrapper

import "testing"

func TestInit(t *testing.T) {
	err := Init("", "")
	if err == nil {
		t.Errorf("Invalid slack connections did not return an error")
	}
}