package fasttui

import (
	"math/rand"
	"strings"
	"testing"
	"unsafe"
)

func TestIsASCII(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"empty", "", true},
		{"single ascii", "a", true},
		{"ascii string", "hello world", true},
		{"ascii with space", "it is a nice day", true},
		{"ascii with punctuation", "Hello, World! 12345", true},
		{"non-ascii chinese", "你好", false},
		{"non-ascii japanese", "こんにちは", false},
		{"non-ascii emoji", "hello 😀", false},
		{"non-ascii mixed", "Hello世界", false},
		{"non-ascii single byte 0x80", "\x80", false},
		{"non-ascii byte 0xFF", "\xff", false},
		{"ascii all codes", " !\"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~", true},
		{"non-ascii DEL 0x7F", "\x7f", true},
		{"long ascii", strings.Repeat("hello world ", 10000), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsASCII(tt.input)
			if got != tt.want {
				t.Errorf("IsASCII(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsPrintableASCII(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"empty", "", true},
		{"single ascii", "a", true},
		{"ascii string", "hello world", true},
		{"ascii with space", "it is a nice day", true},
		{"non-printable tab", "\t", false},
		{"non-printable newline", "\n", false},
		{"non-printable ESC", "\x1b", false},
		{"non-printable mix", "hello\x00world", false},
		{"non-ascii chinese", "你好", false},
		{"non-ascii DEL 0x7F", "\x7f", false},
		{"ascii all printable", " !\"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~", true},
		{"long printable", strings.Repeat("hello world ", 10000), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isPrintableASCII(tt.input)
			if got != tt.want {
				t.Errorf("isPrintableASCII(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// ---- Naive implementations for comparison benchmarking ----

func isASCIINaive(s string) bool {
	for i := range s {
		if s[i] >= 0x80 {
			return false
		}
	}
	return true
}

func isPrintableASCIINaive(s string) bool {
	for i := 0; i < len(s); i++ {
		b := s[i]
		if b < 0x20 || b > 0x7e {
			return false
		}
	}
	return true
}

// ---- Benchmarks ----

func getRandomASCIIString(strLen int) string {
	buf := make([]byte, strLen)
	for i := 0; i < strLen; i++ {
		buf[i] = byte(rand.Intn(128))
	}
	return unsafe.String(&buf[0], strLen)
}

func getRandomPrintableASCIIString(strLen int) string {
	buf := make([]byte, strLen)
	for i := range strLen {
		buf[i] = byte(rand.Intn(95) + 32)
	}
	return unsafe.String(&buf[0], strLen)
}

func BenchmarkIsASCII_naive(b *testing.B) {
	strLen := 1024 * 1024
	s := getRandomASCIIString(strLen)
	s = s[3:]
	b.SetBytes(int64(len(s)))

	for b.Loop() {
		if !isASCIINaive(s) {
			b.Fatal("unexpected non-ASCII")
		}
	}
}

func BenchmarkIsASCII_optimized(b *testing.B) {
	strLen := 1024 * 1024
	s := getRandomASCIIString(strLen)
	s = s[3:]
	b.SetBytes(int64(len(s)))

	for b.Loop() {
		if !IsASCII(s) {
			b.Fatal("unexpected non-ASCII")
		}
	}
}

func BenchmarkIsASCII_naive_short(b *testing.B) {
	s := "hello world this is a typical log message"
	b.SetBytes(int64(len(s)))

	for b.Loop() {
		if !isASCIINaive(s) {
			b.Fatal("unexpected non-ASCII")
		}
	}
}

func BenchmarkIsASCII_optimized_short(b *testing.B) {
	s := "hello world this is a typical log message"
	b.SetBytes(int64(len(s)))

	for b.Loop() {
		if !IsASCII(s) {
			b.Fatal("unexpected non-ASCII")
		}
	}
}

func BenchmarkIsPrintableASCII_naive(b *testing.B) {
	strLen := 1024 * 1024
	s := getRandomPrintableASCIIString(strLen)
	s = s[3:]
	b.SetBytes(int64(len(s)))

	for b.Loop() {
		if !isPrintableASCIINaive(s) {
			b.Fatal("unexpected non-printable ASCII")
		}
	}
}

func BenchmarkIsPrintableASCII_optimized(b *testing.B) {
	strLen := 1024 * 1024
	s := getRandomPrintableASCIIString(strLen)
	s = s[3:]
	b.SetBytes(int64(len(s)))

	for b.Loop() {
		if !isPrintableASCII(s) {
			b.Fatal("unexpected non-printable ASCII")
		}
	}
}

func BenchmarkVisibleWidth_ascii(b *testing.B) {
	s := strings.Repeat("hello world ", 1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = VisibleWidth(s)
	}
}

func BenchmarkVisibleWidth_nonascii(b *testing.B) {
	s := strings.Repeat("Hello世界 ", 1000)

	for b.Loop() {
		_ = VisibleWidth(s)
	}
}
