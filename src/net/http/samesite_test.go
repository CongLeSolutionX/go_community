package http

import "testing"

func TestClientIsSameSiteIncompatible(t *testing.T) {
	tests := []struct {
		name      string
		useragent string
		want      bool
	}{
		// Add test cases.
		{
			name:      "osx 10.14 + safari",
			useragent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_6) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.0.4 Safari/605.1.15",
			want:      true,
		},
		{
			name:      "osx 10.14 + embedded",
			useragent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_4) AppleWebKit/537.36 (KHTML, like Gecko)",
			want:      true,
		},
		{
			name:      "ios 12",
			useragent: "Mozilla/5.0 (iPhone; CPU iPhone OS 12_4 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) CriOS/79.0.3945.73 Mobile/15E148 Safari/605.1",
			want:      true,
		},
		{
			name:      "ucbrowser < 12.13.2",
			useragent: "Mozilla/5.0 (Linux; U; Android 6.0.1; en-US; CPH1701 Build/MMB29M) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/57.0.2987.108 UCBrowser/12.13.0.1207 Mobile Safari/537.36",
			want:      true,
		},
		{
			name:      "chromium based at least version 51",
			useragent: "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Ubuntu Chromium/51.0.2704.106 Chrome/51.0.2704.106 Safari/537.36",
			want:      true,
		},
		{
			name:      "chromium based less than version 67",
			useragent: "Mozilla/5.0 (X11; Linux i686) AppleWebKit/537.36 (KHTML, like Gecko) Ubuntu Chromium/67.0.3396.18 Chrome/67.0.3396.18 Safari/537.36",
			want:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ClientIsSameSiteIncompatible(tt.useragent); got != tt.want {
				t.Errorf("ClientIsSameSiteIncompatible() = %v, want %v", got, tt.want)
			}
		})
	}
}
