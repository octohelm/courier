package jsonflags

import (
	"unicode"
	"unicode/utf8"
)

func foldName(in []byte) []byte {
	// This is inlinable to take advantage of "function outlining".
	// See https://blog.filippo.io/efficient-go-apis-with-the-inliner/
	var arr [32]byte // large enough for most JSON names
	return appendFoldedName(arr[:0], in)
}

func appendFoldedName(out, in []byte) []byte {
	for i := 0; i < len(in); {
		// Handle single-byte ASCII.
		if c := in[i]; c < utf8.RuneSelf {
			if c != '_' && c != '-' {
				if 'a' <= c && c <= 'z' {
					c -= 'a' - 'A'
				}
				out = append(out, c)
			}
			i++
			continue
		}
		// Handle multi-byte Unicode.
		r, n := utf8.DecodeRune(in[i:])
		out = utf8.AppendRune(out, foldRune(r))
		i += n
	}
	return out
}

func foldRune(r rune) rune {
	for {
		r2 := unicode.SimpleFold(r)
		if r2 <= r {
			return r2 // smallest character in the fold set
		}
		r = r2
	}
}
