package cleaner

import (
	"strings"
	"unicode/utf8"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
)

// DecodeZipFilename decodes a raw ZIP filename string to a valid UTF-8 string.
//
// In PKWARE ZIP format, if Bit 11 is not set (nonUTF8 is true), filenames
// are encoded in OEM/local code pages. On Simplified Chinese Windows, this
// is GBK / GB18030 / CP936; on Traditional Chinese Windows it's Big5;
// on Japanese Windows it's Shift-JIS.
//
// Even if nonUTF8 flag was incorrectly omitted by some archivers, if the
// filename contains invalid UTF-8 bytes, we automatically detect and decode it.
func DecodeZipFilename(raw string, nonUTF8 bool) string {
	if raw == "" {
		return ""
	}

	// 1. If explicitly marked as UTF-8 (nonUTF8 == false) and is valid UTF-8, keep it.
	if !nonUTF8 && utf8.ValidString(raw) {
		return raw
	}

	// 2. If it's pure ASCII, all character sets (ASCII, UTF-8, GBK, CP437) are identical.
	if isPureASCII(raw) {
		return raw
	}

	// 3. If nonUTF8 is true, or if raw is NOT valid UTF-8:
	// Prioritize Simplified Chinese (GB18030 / GBK / CP936).
	if decoded, ok := tryDecode(raw, simplifiedchinese.GB18030); ok {
		return decoded
	}

	// 4. Try Traditional Chinese (Big5).
	if decoded, ok := tryDecode(raw, traditionalchinese.Big5); ok {
		return decoded
	}

	// 5. Try Japanese (Shift-JIS).
	if decoded, ok := tryDecode(raw, japanese.ShiftJIS); ok {
		return decoded
	}

	// 6. Fallback: if raw is valid UTF-8, return raw.
	if utf8.ValidString(raw) {
		return raw
	}

	return raw
}

func isPureASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= 0x80 {
			return false
		}
	}
	return true
}

func tryDecode(s string, enc encoding.Encoding) (string, bool) {
	decoded, err := enc.NewDecoder().String(s)
	if err == nil && utf8.ValidString(decoded) && !strings.ContainsRune(decoded, utf8.RuneError) {
		return decoded, true
	}
	return "", false
}
