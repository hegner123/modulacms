package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUserAgentParsing(t *testing.T) {
	tests := []struct {
		name           string
		ua             string
		wantBrowser    string
		wantBrowserVer string
		wantOS         string
		wantOSVer      string
		wantDevice     DeviceType
		wantBot        bool
		wantBotName    string
	}{
		{
			name:           "Chrome on Windows",
			ua:             "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.6778.86 Safari/537.36",
			wantBrowser:    "Chrome",
			wantBrowserVer: "131.0.6778.86",
			wantOS:         "Windows",
			wantOSVer:      "10.0",
			wantDevice:     DeviceDesktop,
		},
		{
			name:           "Firefox on macOS",
			ua:             "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Firefox/133.0",
			wantBrowser:    "Firefox",
			wantBrowserVer: "133.0",
			wantOS:         "macOS",
			wantOSVer:      "10.15.7",
			wantDevice:     DeviceDesktop,
		},
		{
			name:           "Safari on macOS",
			ua:             "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.2 Safari/605.1.15",
			wantBrowser:    "Safari",
			wantBrowserVer: "17.2",
			wantOS:         "macOS",
			wantOSVer:      "10.15.7",
			wantDevice:     DeviceDesktop,
		},
		{
			name:           "Edge on Windows",
			ua:             "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36 Edg/131.0.0.0",
			wantBrowser:    "Edge",
			wantBrowserVer: "131.0.0.0",
			wantOS:         "Windows",
			wantOSVer:      "10.0",
			wantDevice:     DeviceDesktop,
		},
		{
			name:           "Opera on Windows",
			ua:             "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36 OPR/114.0.0.0",
			wantBrowser:    "Opera",
			wantBrowserVer: "114.0.0.0",
			wantOS:         "Windows",
			wantOSVer:      "10.0",
			wantDevice:     DeviceDesktop,
		},
		{
			name:           "Chrome on Android mobile",
			ua:             "Mozilla/5.0 (Linux; Android 14; Pixel 8) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.6778.86 Mobile Safari/537.36",
			wantBrowser:    "Chrome",
			wantBrowserVer: "131.0.6778.86",
			wantOS:         "Android",
			wantOSVer:      "14",
			wantDevice:     DeviceMobile,
		},
		{
			name:           "Safari on iOS",
			ua:             "Mozilla/5.0 (iPhone; CPU iPhone OS 17_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.2 Mobile/15E148 Safari/604.1",
			wantBrowser:    "Safari",
			wantBrowserVer: "17.2",
			wantOS:         "iOS",
			wantOSVer:      "17.2",
			wantDevice:     DeviceMobile,
		},
		{
			name:       "Safari on iPad",
			ua:         "Mozilla/5.0 (iPad; CPU OS 17_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.2 Mobile/15E148 Safari/604.1",
			wantBrowser: "Safari",
			wantBrowserVer: "17.2",
			wantOS:     "iOS",
			wantDevice: DeviceTablet,
		},
		{
			name:           "Chrome on Linux",
			ua:             "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
			wantBrowser:    "Chrome",
			wantBrowserVer: "131.0.0.0",
			wantOS:         "Linux",
			wantDevice:     DeviceDesktop,
		},
		{
			name:           "Chrome on ChromeOS",
			ua:             "Mozilla/5.0 (X11; CrOS x86_64 14541.0.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
			wantBrowser:    "Chrome",
			wantBrowserVer: "131.0.0.0",
			wantOS:         "ChromeOS",
			wantDevice:     DeviceDesktop,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := parseUserAgent(tt.ua)
			if info.BrowserName != tt.wantBrowser {
				t.Errorf("BrowserName = %q, want %q", info.BrowserName, tt.wantBrowser)
			}
			if tt.wantBrowserVer != "" && info.BrowserVersion != tt.wantBrowserVer {
				t.Errorf("BrowserVersion = %q, want %q", info.BrowserVersion, tt.wantBrowserVer)
			}
			if info.OSName != tt.wantOS {
				t.Errorf("OSName = %q, want %q", info.OSName, tt.wantOS)
			}
			if tt.wantOSVer != "" && info.OSVersion != tt.wantOSVer {
				t.Errorf("OSVersion = %q, want %q", info.OSVersion, tt.wantOSVer)
			}
			if info.Device != tt.wantDevice {
				t.Errorf("Device = %q, want %q", info.Device, tt.wantDevice)
			}
			if info.IsBot != tt.wantBot {
				t.Errorf("IsBot = %v, want %v", info.IsBot, tt.wantBot)
			}
			if tt.wantBotName != "" && info.BotName != tt.wantBotName {
				t.Errorf("BotName = %q, want %q", info.BotName, tt.wantBotName)
			}
			if info.Raw != tt.ua {
				t.Errorf("Raw = %q, want %q", info.Raw, tt.ua)
			}
		})
	}
}

func TestUserAgentBotDetection(t *testing.T) {
	tests := []struct {
		name        string
		ua          string
		wantBot     bool
		wantBotName string
	}{
		{"Googlebot", "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)", true, "Googlebot"},
		{"Bingbot", "Mozilla/5.0 (compatible; bingbot/2.0; +http://www.bing.com/bingbot.htm)", true, "Bingbot"},
		{"Baiduspider", "Mozilla/5.0 (compatible; Baiduspider/2.0; +http://www.baidu.com/search/spider.html)", true, "Baiduspider"},
		{"YandexBot", "Mozilla/5.0 (compatible; YandexBot/3.0; +http://yandex.com/bots)", true, "YandexBot"},
		{"DuckDuckBot", "DuckDuckBot/1.1; (+http://duckduckgo.com/duckduckbot.html)", true, "DuckDuckBot"},
		{"Slurp", "Mozilla/5.0 (compatible; Yahoo! Slurp; http://help.yahoo.com/help/us/ysearch/slurp)", true, "Yahoo! Slurp"},
		{"Facebook", "facebookexternalhit/1.1 (+http://www.facebook.com/externalhit_uatext.php)", true, "Facebook"},
		{"Twitterbot", "Twitterbot/1.0", true, "Twitterbot"},
		{"LinkedInBot", "LinkedInBot/1.0 (compatible; Mozilla/5.0; Apache-HttpClient +http://www.linkedin.com)", true, "LinkedInBot"},
		{"generic bot", "MyCustomBot/1.0", true, "generic"},
		{"generic crawler", "CustomCrawler/2.0", true, "generic"},
		{"generic spider", "WebSpider/3.0", true, "generic"},
		{"empty UA", "", true, "unknown"},
		{"normal browser", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/131.0.0.0 Safari/537.36", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := parseUserAgent(tt.ua)
			if info.IsBot != tt.wantBot {
				t.Errorf("IsBot = %v, want %v", info.IsBot, tt.wantBot)
			}
			if info.BotName != tt.wantBotName {
				t.Errorf("BotName = %q, want %q", info.BotName, tt.wantBotName)
			}
			if tt.wantBot && info.Device != DeviceBot {
				t.Errorf("Device = %q, want %q for bot", info.Device, DeviceBot)
			}
		})
	}
}

func TestParseSecCHUA(t *testing.T) {
	tests := []struct {
		name       string
		header     string
		wantBrands []BrandVersion
	}{
		{
			name:   "Chrome with GREASE",
			header: `"Chromium";v="131", "Google Chrome";v="131", "Not_A Brand";v="24"`,
			wantBrands: []BrandVersion{
				{Brand: "Chromium", Version: "131"},
				{Brand: "Google Chrome", Version: "131"},
			},
		},
		{
			name:   "Edge with GREASE",
			header: `"Not A(Brand";v="99", "Microsoft Edge";v="131", "Chromium";v="131"`,
			wantBrands: []BrandVersion{
				{Brand: "Microsoft Edge", Version: "131"},
				{Brand: "Chromium", Version: "131"},
			},
		},
		{
			name:       "empty header",
			header:     "",
			wantBrands: nil,
		},
		{
			name:   "single brand no GREASE",
			header: `"Firefox";v="133"`,
			wantBrands: []BrandVersion{
				{Brand: "Firefox", Version: "133"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			brands := parseSecCHUA(tt.header)
			if len(brands) != len(tt.wantBrands) {
				t.Fatalf("got %d brands, want %d: %+v", len(brands), len(tt.wantBrands), brands)
			}
			for i, want := range tt.wantBrands {
				if brands[i].Brand != want.Brand {
					t.Errorf("brands[%d].Brand = %q, want %q", i, brands[i].Brand, want.Brand)
				}
				if brands[i].Version != want.Version {
					t.Errorf("brands[%d].Version = %q, want %q", i, brands[i].Version, want.Version)
				}
			}
		})
	}
}

func TestParseClientHints(t *testing.T) {
	t.Run("all hints present", func(t *testing.T) {
		h := http.Header{}
		h.Set("Sec-CH-UA", `"Chromium";v="131", "Google Chrome";v="131", "Not_A Brand";v="24"`)
		h.Set("Sec-CH-UA-Mobile", "?1")
		h.Set("Sec-CH-UA-Platform", `"Android"`)
		h.Set("Sec-CH-UA-Platform-Version", `"14.0"`)
		h.Set("Sec-CH-UA-Model", `"Pixel 8"`)
		h.Set("Sec-CH-UA-Arch", `"arm"`)

		var info UserAgentInfo
		parseClientHints(h, &info)

		if !info.HintsAvailable {
			t.Error("HintsAvailable = false, want true")
		}
		if len(info.Brands) != 2 {
			t.Fatalf("got %d brands, want 2", len(info.Brands))
		}
		if !info.Mobile {
			t.Error("Mobile = false, want true")
		}
		if info.Platform != "Android" {
			t.Errorf("Platform = %q, want %q", info.Platform, "Android")
		}
		if info.PlatformVersion != "14.0" {
			t.Errorf("PlatformVersion = %q, want %q", info.PlatformVersion, "14.0")
		}
		if info.Model != "Pixel 8" {
			t.Errorf("Model = %q, want %q", info.Model, "Pixel 8")
		}
		if info.Architecture != "arm" {
			t.Errorf("Architecture = %q, want %q", info.Architecture, "arm")
		}
	})

	t.Run("no hints", func(t *testing.T) {
		h := http.Header{}
		var info UserAgentInfo
		parseClientHints(h, &info)

		if info.HintsAvailable {
			t.Error("HintsAvailable = true, want false")
		}
	})

	t.Run("mobile not set", func(t *testing.T) {
		h := http.Header{}
		h.Set("Sec-CH-UA-Mobile", "?0")
		h.Set("Sec-CH-UA-Platform", `"Windows"`)

		var info UserAgentInfo
		parseClientHints(h, &info)

		if info.Mobile {
			t.Error("Mobile = true, want false")
		}
		if info.Platform != "Windows" {
			t.Errorf("Platform = %q, want %q", info.Platform, "Windows")
		}
	})
}

func TestUserAgentMiddleware(t *testing.T) {
	handler := UserAgentMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		info := UserAgentInfoFromContext(r.Context())
		if info.BrowserName != "Chrome" {
			t.Errorf("BrowserName = %q, want %q", info.BrowserName, "Chrome")
		}
		if info.OSName != "Windows" {
			t.Errorf("OSName = %q, want %q", info.OSName, "Windows")
		}
		if info.Device != DeviceDesktop {
			t.Errorf("Device = %q, want %q", info.Device, DeviceDesktop)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/",
		nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	acceptCH := rec.Header().Get("Accept-CH")
	if acceptCH == "" {
		t.Error("Accept-CH header not set")
	}
	if acceptCH != acceptCHValue {
		t.Errorf("Accept-CH = %q, want %q", acceptCH, acceptCHValue)
	}
}

func TestUserAgentInfoFromContext_Missing(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	info := UserAgentInfoFromContext(req.Context())

	if info.Raw != "" {
		t.Errorf("Raw = %q, want empty", info.Raw)
	}
	if info.BrowserName != "" {
		t.Errorf("BrowserName = %q, want empty", info.BrowserName)
	}
	if info.Device != "" {
		t.Errorf("Device = %q, want empty", info.Device)
	}
	if info.IsBot {
		t.Error("IsBot = true, want false")
	}
}

func TestExtractVersionAfter(t *testing.T) {
	tests := []struct {
		name   string
		ua     string
		prefix string
		want   string
	}{
		{"simple version", "Chrome/131.0", "Chrome/", "131.0"},
		{"full version", "Chrome/131.0.6778.86 Safari/537.36", "Chrome/", "131.0.6778.86"},
		{"not found", "Firefox/133.0", "Chrome/", ""},
		{"version with underscore", "Mac OS X 10_15_7)", "Mac OS X ", "10_15_7"},
		{"version at end", "Firefox/133.0", "Firefox/", "133.0"},
		{"trailing dot stripped", "Chrome/131.", "Chrome/", "131"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractVersionAfter(tt.ua, tt.prefix)
			if got != tt.want {
				t.Errorf("extractVersionAfter(%q, %q) = %q, want %q", tt.ua, tt.prefix, got, tt.want)
			}
		})
	}
}

func TestUnquoteHeader(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{`"Windows"`, "Windows"},
		{`"Android"`, "Android"},
		{`Windows`, "Windows"},
		{`""`, ""},
		{`"x"`, "x"},
		{`"`, `"`},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := unquoteHeader(tt.input)
			if got != tt.want {
				t.Errorf("unquoteHeader(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
