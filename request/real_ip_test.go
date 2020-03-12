package request

import (
	"context"
	"net/http"
	"testing"

	"github.com/raspberry-gateway/raspberry/headers"
)

var ipHeaderTests = []struct {
	remoteAddr string
	key        string
	value      string
	expected   string
	comment    string
}{
	{remoteAddr: "192.168.1.10:8080", key: "X-Real-IP", value: "192.168.1.1", expected: "192.168.1.1", comment: "X-Real-IP"},
	{remoteAddr: "192.168.1.10:8080", key: "X-Forwarded-For", value: "192.168.1.2", expected: "192.168.1.2", comment: "X-Forwarded-For"},
	{remoteAddr: "192.168.1.10:8080", key: "X-Forwarded-For", value: "192.168.1.3, 192.168.1.2, 192.168.1.1", expected: "192.168.1.3", comment: "X-Forwarded-For (multiple)"},
	{remoteAddr: "192.168.1.10:8080", expected: "192.168.1.10", comment: "RemoteAddr"},
}

var testURL = "http://xyz.com"

func TestRealIP(t *testing.T) {
	for _, test := range ipHeaderTests {
		t.Log(test.comment)

		r, _ := http.NewRequest(http.MethodGet, testURL+":8080", nil)
		r.Header.Set(test.key, test.value)
		r.RemoteAddr = test.remoteAddr

		ip := RealIP(r)

		if ip != test.expected {
			t.Errorf("\texpected %s got %s", test.expected, ip)
		}
	}

	t.Log("Context")
	r, _ := http.NewRequest(http.MethodGet, testURL+":8080", nil)

	ctx := context.WithValue(r.Context(), headers.RemoteAddr, "192.168.0.5")
	r = r.WithContext(ctx)

	ip := RealIP(r)
	if ip != "192.168.0.5" {
		t.Errorf("\texpected %s got %s", "192.168.0.5", ip)
	}
}

func BenchmarkRealIP_RemoteAddr(b *testing.B) {
	b.ReportAllocs()

	r, _ := http.NewRequest(http.MethodGet, testURL+":8080", nil)
	r.RemoteAddr = testURL + ":8081"

	for n := 0; n < b.N; n++ {
		ip := RealIP(r)
		if ip != "192.168.1.10" {
			b.Errorf("\texpected %s got %s", "192.168.1.10", ip)
		}
	}
}

func BenchmarkRealIP_ForwardedFor(b *testing.B) {
	b.ReportAllocs()

	r, _ := http.NewRequest(http.MethodGet, testURL+":8080", nil)
	r.Header.Set(headers.XForwardFor, "192.168.1.3, 192.168.1.2, 192.168.1.1")

	for n := 0; n < b.N; n++ {
		ip := RealIP(r)
		if ip != "192.168.1.3" {
			b.Errorf("\texpected %s got %s", "192.168.1.3", ip)
		}
	}
}

func BenchmarkRealIP_RealIP(b *testing.B) {
	b.ReportAllocs()

	r, _ := http.NewRequest(http.MethodGet, testURL+":8080", nil)
	r.Header.Set(headers.XRealIP, "192.168.1.10")

	for n := 0; n < b.N; n++ {
		ip := RealIP(r)
		if ip != "192.168.1.10" {
			b.Errorf("\texpected %s got %s", "192.168.1.10", ip)
		}
	}
}

func BenchmarkRealIP_Context(b *testing.B) {
	b.ReportAllocs()

	r, _ := http.NewRequest(http.MethodGet, testURL+":8080", nil)
	r.Header.Set(headers.XRealIP, "192.168.1.10")
	ctx := context.WithValue(r.Context(), headers.RemoteAddr, "192.168.1.5")
	r = r.WithContext(ctx)

	for n := 0; n < b.N; n++ {
		ip := RealIP(r)
		if ip != "192.168.1.5" {
			b.Errorf("\texpected %s got %s", "192.168.1.5", ip)
		}
	}
}
