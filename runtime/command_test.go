package runtime

import "testing"

func TestTokenizer_dash(t *testing.T) {
	expected := 1
	segments := tokenize("hello-my-friend")
	if len(segments) != expected {
		t.Errorf("Invalid segment count %d vs. %d: %+v", expected, len(segments), segments)
	}
}

func TestTokenizer_space(t *testing.T) {
	expected := 3
	segments := tokenize("hello my friend")
	if len(segments) != expected {
		t.Errorf("Invalid segment count %d vs. %d", expected, len(segments))
	}
}
