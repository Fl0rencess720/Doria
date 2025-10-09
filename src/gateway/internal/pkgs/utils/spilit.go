package utils

import (
	"bytes"
	"unicode/utf8"
)

var (
	ellipsisASCII   = []byte("...")
	ellipsisUnicode = []byte("…")

	singlePunctuationDelimiters = []struct {
		r    rune
		size int
	}{
		{',', utf8.RuneLen(',')},
		{'.', utf8.RuneLen('.')},
		{'!', utf8.RuneLen('!')},
		{'?', utf8.RuneLen('?')},
		{';', utf8.RuneLen(';')},
		{'，', utf8.RuneLen('，')},
		{'。', utf8.RuneLen('。')},
		{'！', utf8.RuneLen('！')},
		{'？', utf8.RuneLen('？')},
		{'；', utf8.RuneLen('；')},
	}
)

func ScanOnPunctuation(data []byte, atEOF bool) (advance int, token []byte, err error) {
	tokenStart := 0
	delimiterIndex := -1
	delimiterLen := 0
	for i := 0; i < len(data); {
		if bytes.HasPrefix(data[i:], ellipsisASCII) {
			delimiterIndex = i
			delimiterLen = len(ellipsisASCII)
			break
		}
		if bytes.HasPrefix(data[i:], ellipsisUnicode) {
			delimiterIndex = i
			delimiterLen = len(ellipsisUnicode)
			break
		}
		r, size := utf8.DecodeRune(data[i:])
		foundSinglePunctuation := false
		for _, p := range singlePunctuationDelimiters {
			if r == p.r {
				delimiterIndex = i
				delimiterLen = size
				foundSinglePunctuation = true
				break
			}
		}
		if foundSinglePunctuation {
			break
		}
		i += size
	}

	if delimiterIndex != -1 {
		return delimiterIndex + delimiterLen, data[tokenStart : delimiterIndex+delimiterLen], nil
	}

	if atEOF && len(data) > 0 {
		return len(data), data[tokenStart:], nil
	}

	return 0, nil, nil
}
