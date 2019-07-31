package runtime

import (
	"testing"
)

func TestParseUsers(t *testing.T) {
	l := `<@U6X4TQW4Q>: men jeg mener ikke vi skal til at staa for installation af noget som helst.. heller ikke selv appen.. vi
skal lave en beskrivelse af hvordan de aar den til at koere :)`
	s := parseUsers(nil, l)
	t.Log(s)
	if s == l {
		t.Fail()
	}
}
