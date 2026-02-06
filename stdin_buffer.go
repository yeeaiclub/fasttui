package fasttui

const (
	ESC                   = "\x1b"
	BRACKETED_PASTE_START = "\x1b[200~"
	BRACKETED_PASTE_END   = "\x1b[201~"
)

// SequenceStatus 转义序列状态类型
type SequenceStatus string

const (
	SequenceComplete   SequenceStatus = "complete"
	SequenceIncomplete SequenceStatus = "incomplete"
	SequenceNotEscape  SequenceStatus = "not-escape"
)

// IsCompleteSequence 检查字符串是否是完整的转义序列或需要更多数据
// @param data - 要检查的字符串
// @returns 序列状态：complete、incomplete 或 not-escape
func IsCompleteSequence(data string) SequenceStatus {
	if !startsWith(data, ESC) {
		return SequenceNotEscape
	}

	if len(data) == 1 {
		return SequenceIncomplete
	}

	afterEsc := data[1:]

	// CSI 序列: ESC [
	if startsWith(afterEsc, "[") {
		// 检查旧式鼠标序列: ESC[M + 3 字节
		if startsWith(afterEsc, "[M") {
			// 旧式鼠标需要 ESC[M + 3 字节 = 总共 6 个
			if len(data) >= 6 {
				return SequenceComplete
			}
			return SequenceIncomplete
		}
		return IsCompleteCsiSequence(data)
	}

	// OSC 序列: ESC ]
	if startsWith(afterEsc, "]") {
		return IsCompleteOscSequence(data)
	}

	// DCS 序列: ESC P ... ESC \ (包括 XTVersion 响应)
	if startsWith(afterEsc, "P") {
		return IsCompleteDcsSequence(data)
	}

	// APC 序列: ESC _ ... ESC \ (包括 Kitty 图形响应)
	if startsWith(afterEsc, "_") {
		return IsCompleteApcSequence(data)
	}

	// SS3 序列: ESC O
	if startsWith(afterEsc, "O") {
		// ESC O 后跟单个字符
		if len(afterEsc) >= 2 {
			return SequenceComplete
		}
		return SequenceIncomplete
	}

	// Meta 键序列: ESC 后跟单个字符
	if len(afterEsc) == 1 {
		return SequenceComplete
	}

	// 未知转义序列 - 视为完整
	return SequenceComplete
}

// IsCompleteCsiSequence 检查 CSI 序列是否完整
// CSI 序列: ESC [ ... 后跟最终字节 (0x40-0x7E)
// @param data - 要检查的字符串
// @returns 序列状态：complete 或 incomplete
func IsCompleteCsiSequence(data string) SequenceStatus {
	if !startsWith(data, ESC+"[") {
		return SequenceComplete
	}

	// 至少需要 ESC [ 和一个以上字符
	if len(data) < 3 {
		return SequenceIncomplete
	}

	payload := data[2:]

	// CSI 序列以 0x40-0x7E 范围内的字节结束
	lastChar := payload[len(payload)-1]
	lastCharCode := int(lastChar)

	if lastCharCode >= 0x40 && lastCharCode <= 0x7E {
		// 特殊处理 SGR 鼠标序列
		// 格式: ESC[<B;X;Ym 或 ESC[<B;X;YM
		if startsWith(payload, "<") {
			// 必须具有格式: <数字;数字;数字[Mm]
			if len(payload) >= 5 {
				lastTwo := payload[len(payload)-2:]
				if (lastTwo[1] == 'M' || lastTwo[1] == 'm') && lastTwo[0] == ';' {
					parts := splitString(payload[1:len(payload)-1], ';')
					if len(parts) == 3 {
						allDigits := true
						for _, part := range parts {
							if !isDigits(part) {
								allDigits = false
								break
							}
						}
						if allDigits {
							return SequenceComplete
						}
					}
				}
			}
			return SequenceIncomplete
		}

		return SequenceComplete
	}

	return SequenceIncomplete
}

// IsCompleteOscSequence 检查 OSC 序列是否完整
// OSC 序列: ESC ] ... ST (其中 ST 是 ESC \ 或 BEL)
// @param data - 要检查的字符串
// @returns 序列状态：complete 或 incomplete
func IsCompleteOscSequence(data string) SequenceStatus {
	if !startsWith(data, ESC+"]") {
		return SequenceComplete
	}

	// OSC 序列以 ST (ESC \) 或 BEL (\x07) 结束
	if endsWith(data, ESC+"\\") || endsWith(data, "\x07") {
		return SequenceComplete
	}

	return SequenceIncomplete
}

// IsCompleteDcsSequence 检查 DCS (设备控制字符串) 序列是否完整
// DCS 序列: ESC P ... ST (其中 ST 是 ESC \)
// 用于 XTVersion 响应，如 ESC P >| ... ESC \
// @param data - 要检查的字符串
// @returns 序列状态：complete 或 incomplete
func IsCompleteDcsSequence(data string) SequenceStatus {
	if !startsWith(data, ESC+"P") {
		return SequenceComplete
	}

	// DCS 序列以 ST (ESC \) 结束
	if endsWith(data, ESC+"\\") {
		return SequenceComplete
	}

	return SequenceIncomplete
}

// IsCompleteApcSequence 检查 APC (应用程序命令) 序列是否完整
// APC 序列: ESC _ ... ST (其中 ST 是 ESC \)
// 用于 Kitty 图形响应，如 ESC _ G ... ESC \
// @param data - 要检查的字符串
// @returns 序列状态：complete 或 incomplete
func IsCompleteApcSequence(data string) SequenceStatus {
	if !startsWith(data, ESC+"_") {
		return SequenceComplete
	}

	// APC 序列以 ST (ESC \) 结束
	if endsWith(data, ESC+"\\") {
		return SequenceComplete
	}

	return SequenceIncomplete
}

// 辅助函数

// startsWith 检查字符串是否以指定前缀开头
func startsWith(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

// endsWith 检查字符串是否以指定后缀结尾
func endsWith(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}

// splitString 分割字符串
func splitString(s string, sep rune) []string {
	var result []string
	var current string
	for _, r := range s {
		if r == sep {
			result = append(result, current)
			current = ""
		} else {
			current += string(r)
		}
	}
	result = append(result, current)
	return result
}

// isDigits 检查字符串是否只包含数字
func isDigits(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return len(s) > 0
}
