package fasttui

import "testing"

func TestIsCompleteSequence(t *testing.T) {
	testCases := []struct {
		name     string
		data     string
		expected SequenceStatus
	}{
		// 非转义序列
		{"Non-escape sequence", "hello", SequenceNotEscape},
		{"Empty string", "", SequenceNotEscape},

		// 不完整的转义序列
		{"Incomplete escape (only ESC)", ESC, SequenceIncomplete},
		{"Incomplete CSI (only ESC[)", ESC + "[", SequenceIncomplete},
		{"Incomplete SS3 (only ESC O)", ESC + "O", SequenceIncomplete},

		// 完整的转义序列
		{"Complete meta key", ESC + "a", SequenceComplete},
		{"Complete SS3", ESC + "OP", SequenceComplete},

		// CSI 序列
		{"Complete CSI (cursor up)", ESC + "[A", SequenceComplete},
		{"Complete CSI (cursor down with param)", ESC + "[5B", SequenceComplete},
		{"Complete CSI (mouse sequence)", ESC + "[M@", SequenceIncomplete}, // 需要更多字节

		// OSC 序列
		{"Incomplete OSC", ESC + "]0;title", SequenceIncomplete},
		{"Complete OSC (with BEL)", ESC + "]0;title\x07", SequenceComplete},
		{"Complete OSC (with ST)", ESC + "]0;title" + ESC + "\\", SequenceComplete},

		// DCS 序列
		{"Incomplete DCS", ESC + "Ptest", SequenceIncomplete},
		{"Complete DCS", ESC + "Ptest" + ESC + "\\", SequenceComplete},

		// APC 序列
		{"Incomplete APC", ESC + "_test", SequenceIncomplete},
		{"Complete APC", ESC + "_test" + ESC + "\\", SequenceComplete},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsCompleteSequence(tc.data)
			if result != tc.expected {
				t.Errorf("Expected %q, got %q for data %q", tc.expected, result, tc.data)
			}
		})
	}
}

func TestIsCompleteCsiSequence(t *testing.T) {
	testCases := []struct {
		name     string
		data     string
		expected SequenceStatus
	}{
		{"Not CSI sequence", "hello", SequenceComplete},
		{"Incomplete CSI (only ESC[)", ESC + "[", SequenceIncomplete},
		{"Complete CSI (cursor up)", ESC + "[A", SequenceComplete},
		{"Complete CSI (cursor down with param)", ESC + "[5B", SequenceComplete},
		{"Incomplete CSI (missing final char)", ESC + "[5", SequenceIncomplete},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsCompleteCsiSequence(tc.data)
			if result != tc.expected {
				t.Errorf("Expected %q, got %q for data %q", tc.expected, result, tc.data)
			}
		})
	}
}

func TestIsCompleteOscSequence(t *testing.T) {
	testCases := []struct {
		name     string
		data     string
		expected SequenceStatus
	}{
		{"Not OSC sequence", "hello", SequenceComplete},
		{"Incomplete OSC (only ESC])", ESC + "]", SequenceIncomplete},
		{"Incomplete OSC (no terminator)", ESC + "]0;title", SequenceIncomplete},
		{"Complete OSC (with BEL)", ESC + "]0;title\x07", SequenceComplete},
		{"Complete OSC (with ST)", ESC + "]0;title" + ESC + "\\", SequenceComplete},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsCompleteOscSequence(tc.data)
			if result != tc.expected {
				t.Errorf("Expected %q, got %q for data %q", tc.expected, result, tc.data)
			}
		})
	}
}

func TestIsCompleteDcsSequence(t *testing.T) {
	testCases := []struct {
		name     string
		data     string
		expected SequenceStatus
	}{
		{"Not DCS sequence", "hello", SequenceComplete},
		{"Incomplete DCS (only ESC P)", ESC + "P", SequenceIncomplete},
		{"Incomplete DCS (no terminator)", ESC + "Ptest", SequenceIncomplete},
		{"Complete DCS", ESC + "Ptest" + ESC + "\\", SequenceComplete},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsCompleteDcsSequence(tc.data)
			if result != tc.expected {
				t.Errorf("Expected %q, got %q for data %q", tc.expected, result, tc.data)
			}
		})
	}
}

func TestIsCompleteApcSequence(t *testing.T) {
	testCases := []struct {
		name     string
		data     string
		expected SequenceStatus
	}{
		{"Not APC sequence", "hello", SequenceComplete},
		{"Incomplete APC (only ESC _)", ESC + "_", SequenceIncomplete},
		{"Incomplete APC (no terminator)", ESC + "_test", SequenceIncomplete},
		{"Complete APC", ESC + "_test" + ESC + "\\", SequenceComplete},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsCompleteApcSequence(tc.data)
			if result != tc.expected {
				t.Errorf("Expected %q, got %q for data %q", tc.expected, result, tc.data)
			}
		})
	}
}

func TestHelperFunctions(t *testing.T) {
	// 测试 startsWith
	if !startsWith("hello", "he") {
		t.Error("startsWith failed for prefix match")
	}
	if startsWith("hello", "world") {
		t.Error("startsWith failed for non-prefix match")
	}

	// 测试 endsWith
	if !endsWith("hello", "lo") {
		t.Error("endsWith failed for suffix match")
	}
	if endsWith("hello", "world") {
		t.Error("endsWith failed for non-suffix match")
	}

	// 测试 splitString
	result := splitString("1;2;3", ';')
	if len(result) != 3 || result[0] != "1" || result[1] != "2" || result[2] != "3" {
		t.Error("splitString failed")
	}

	// 测试 isDigits
	if !isDigits("123") {
		t.Error("isDigits failed for digits")
	}
	if isDigits("abc") {
		t.Error("isDigits failed for non-digits")
	}
	if isDigits("") {
		t.Error("isDigits failed for empty string")
	}
}
