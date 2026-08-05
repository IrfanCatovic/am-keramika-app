package utils

import "testing"

func TestGenerateSlug(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Keramika", "keramika"},
		{"Podne pločice", "podne-plocice"},
		{"  Široki  razmak  ", "siroki-razmak"},
		{"Đakovo", "djakovo"},
		{"", ""},
	}

	for _, tt := range tests {
		got := GenerateSlug(tt.input)
		if got != tt.want {
			t.Errorf("GenerateSlug(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
