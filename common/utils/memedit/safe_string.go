package memedit

import (
	"unicode/utf8"

	"github.com/yaklang/yaklang/common/yak/yaklib/codec"
)

type SafeString struct {
	utf8Valid bool

	runes []rune
	bytes []byte
}

func NewSafeString(i any) *SafeString {
	ss := &SafeString{}
	raw := codec.AnyToBytes(i)
	if utf8.Valid(raw) {
		ss.utf8Valid = true
		ss.runes = []rune(string(raw))
	} else {
		ss.bytes = raw
	}
	return ss
}

func (s *SafeString) SafeSlice2(start, end int) *SafeString {
	if s.utf8Valid {
		return &SafeString{
			utf8Valid: s.utf8Valid,
			runes:     s.runes[start:end],
		}
	}
	return &SafeString{
		utf8Valid: s.utf8Valid,
		bytes:     s.bytes[start:end],
	}
}

func (s *SafeString) Slice2(start, end int) string {
	if s.utf8Valid {
		return string(s.runes[start:end])
	}
	return string(s.bytes[start:end])
}

func (s *SafeString) SafeSliceToEnd(start int) *SafeString {
	if s.utf8Valid {
		return &SafeString{
			utf8Valid: s.utf8Valid,
			runes:     s.runes[start:],
		}
	}
	return &SafeString{
		utf8Valid: s.utf8Valid,
		bytes:     s.bytes[start:],
	}
}

func (s *SafeString) SliceToEnd(start int) string {
	if s.utf8Valid {
		return string(s.runes[start:])
	}
	return string(s.bytes[start:])
}

func (s *SafeString) SliceBeforeStart(end int) string {
	if s.utf8Valid {
		return string(append([]rune{}, s.runes[:end]...))
	}
	return string(append([]byte{}, s.bytes[:end]...))
}

func (s *SafeString) Slice1(idx int) rune {
	if idx < 0 {
		return 0
	}

	if idx >= s.Len() {
		return 0
	}

	if s.utf8Valid {
		return s.runes[idx]
	}
	return rune(s.bytes[idx])
}

func (s *SafeString) Runes() []rune {
	if s.utf8Valid {
		return s.runes
	}
	return []rune(string(s.bytes))
}

func (s *SafeString) Bytes() []byte {
	if s.utf8Valid {
		return []byte(string(s.runes))
	}
	return s.bytes
}

func (s *SafeString) String() string {
	if s.utf8Valid {
		return string(s.runes)
	}
	return string(s.bytes)
}

func (s *SafeString) Len() int {
	if s.utf8Valid {
		return len(s.runes)
	}
	return len(s.bytes)
}

func (s *SafeString) IndexString(what string) int {
	return s.Index([]rune(what))
}

func (s *SafeString) Index(what []rune) int {
	if len(what) == 0 {
		return 0
	}
	if len(what) > len(s.runes) {
		return -1
	}

	// 使用KMP算法匹配字符串
	// 构建next数组
	next := make([]int, len(what))
	next[0] = -1
	i, j := 0, -1
	for i < len(what)-1 {
		if j == -1 || what[i] == what[j] {
			i++
			j++
			next[i] = j
		} else {
			j = next[j]
		}
	}

	// 搜索
	i, j = 0, 0
	for i < len(s.runes) && j < len(what) {
		if j == -1 || s.runes[i] == what[j] {
			i++
			j++
		} else {
			j = next[j]
		}
	}
	if j == len(what) {
		return i - j
	}
	return -1
}
