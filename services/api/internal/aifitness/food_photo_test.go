package aifitness

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestValidateFoodImageDataURL(t *testing.T) {
	valid := "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString([]byte("image bytes"))
	if err := validateFoodImageDataURL(valid); err != nil {
		t.Fatalf("valid image rejected: %v", err)
	}

	tests := []struct {
		name  string
		value string
		want  string
	}{
		{name: "not data url", value: "https://example.com/photo.jpg", want: "data URL"},
		{name: "unsupported mime", value: "data:image/gif;base64,R0lGODlh", want: "JPEG, PNG, or WebP"},
		{name: "empty", value: "data:image/png;base64,", want: "empty"},
		{name: "invalid base64", value: "data:image/webp;base64,%%%", want: "invalid base64"},
		{name: "oversize", value: "data:image/jpeg;base64," + strings.Repeat("A", 7_000_000), want: "5 MB"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateFoodImageDataURL(test.value)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("got %v, want error containing %q", err, test.want)
			}
		})
	}
}
