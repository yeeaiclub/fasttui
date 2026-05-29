package fasttui

import (
	"unsafe"
)

const maskOfAscii uint64 = 0x8080808080808080

func IsASCII(s string) bool {
	l := len(s)
	if l == 0 {
		return true
	}

	addr := uint64(uintptr(unsafe.Pointer(unsafe.StringData(s))))
	alignAddr := (addr + 63) &^ 63
	headLen := int(alignAddr - addr)
	if headLen > l {
		headLen = l
	}

	ptr := unsafe.Pointer(unsafe.StringData(s))

	for i := 0; i < headLen; i++ {
		if *(*byte)(unsafe.Add(ptr, i)) >= 0x80 {
			return false
		}
	}

	remaining := l - headLen
	align64 := remaining &^ 63
	ptr = unsafe.Add(ptr, headLen)

	for i := 0; i < align64; i += 64 {
		values := (*[8]uint64)(unsafe.Add(ptr, i))
		if (values[0]|values[1]|values[2]|values[3]|
			values[4]|values[5]|values[6]|values[7])&maskOfAscii != 0 {
			return false
		}
	}

	ptr = unsafe.Add(ptr, align64)
	tailLen := remaining & 63
	for i := 0; i < tailLen; i++ {
		if *(*byte)(unsafe.Add(ptr, i)) >= 0x80 {
			return false
		}
	}

	return true
}

func isPrintableASCII(s string) bool {
	l := len(s)
	if l == 0 {
		return true
	}

	addr := uint64(uintptr(unsafe.Pointer(unsafe.StringData(s))))
	alignAddr := (addr + 63) &^ 63
	headLen := int(alignAddr - addr)
	if headLen > l {
		headLen = l
	}

	ptr := unsafe.Pointer(unsafe.StringData(s))

	for i := 0; i < headLen; i++ {
		b := *(*byte)(unsafe.Add(ptr, i))
		if b < 0x20 || b > 0x7e {
			return false
		}
	}

	remaining := l - headLen
	align64 := remaining &^ 63
	ptr = unsafe.Add(ptr, headLen)

	const maskLowerBound uint64 = 0xE0E0E0E0E0E0E0E0

	for i := 0; i < align64; i += 64 {
		values := (*[8]uint64)(unsafe.Add(ptr, i))
		v0, v1 := values[0], values[1]
		v2, v3 := values[2], values[3]
		v4, v5 := values[4], values[5]
		v6, v7 := values[6], values[7]

		combined := v0 | v1 | v2 | v3 | v4 | v5 | v6 | v7
		if combined&maskOfAscii != 0 {
			return false
		}

		lower := (v0 + maskLowerBound) | (v1 + maskLowerBound) |
			(v2 + maskLowerBound) | (v3 + maskLowerBound) |
			(v4 + maskLowerBound) | (v5 + maskLowerBound) |
			(v6 + maskLowerBound) | (v7 + maskLowerBound)
		if lower&maskOfAscii != 0 {
			return false
		}
	}

	ptr = unsafe.Add(ptr, align64)
	tailLen := remaining & 63
	for i := range tailLen {
		b := *(*byte)(unsafe.Add(ptr, i))
		if b < 0x20 || b > 0x7e {
			return false
		}
	}

	return true
}

// IsPrintableASCII reports whether s contains only printable ASCII (0x20-0x7e).
func IsPrintableASCII(s string) bool {
	return isPrintableASCII(s)
}
