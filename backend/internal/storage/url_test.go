package storage

import (
	"testing"
)

func TestResolveMediaURL(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		s3ObjectURL string
		forSurface  string
		expected    string
	}{
		{
			name:        "local public api default",
			key:         "media/abc-123.jpg",
			s3ObjectURL: "",
			forSurface:  "",
			expected:    "/api/v1/media/media/abc-123.jpg",
		},
		{
			name:        "local dashboard api surface",
			key:         "media/abc-123.jpg",
			s3ObjectURL: "",
			forSurface:  "dashboard",
			expected:    "/dashboard/api/media/media/abc-123.jpg",
		},
		{
			name:        "s3 configured without trailing slash",
			key:         "media/abc-123.jpg",
			s3ObjectURL: "https://my-bucket.s3.amazonaws.com",
			forSurface:  "dashboard",
			expected:    "https://my-bucket.s3.amazonaws.com/media/abc-123.jpg",
		},
		{
			name:        "s3 configured with trailing slash and leading key slash",
			key:         "/media/abc-123.jpg",
			s3ObjectURL: "https://my-bucket.s3.amazonaws.com/",
			forSurface:  "",
			expected:    "https://my-bucket.s3.amazonaws.com/media/abc-123.jpg",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ResolveMediaURL(tc.key, tc.s3ObjectURL, tc.forSurface)
			if got != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, got)
			}
		})
	}
}
