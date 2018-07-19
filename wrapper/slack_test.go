package wrapper

import "testing"

func TestInit(t *testing.T) {
	_, err := New("", "")
	if err == nil {
		t.Errorf("Invalid slack connections did not return an error")
	}
}
