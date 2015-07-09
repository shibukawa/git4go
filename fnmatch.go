package git4go

import (
	"bytes"
	"errors"
	"strings"
)

type FnMatchFlag int
type RangeMatchResult int

const (
	FNMNoEscape   FnMatchFlag = 1 << iota
	FNMPathName   FnMatchFlag = 1 << iota
	FNMPeriod     FnMatchFlag = 1 << iota
	FNMLeadingDir FnMatchFlag = 1 << iota
	FNMCaseFold   FnMatchFlag = 1 << iota

	RangeMatch   RangeMatchResult = 1
	RangeNoMatch RangeMatchResult = 0
	RangeError   RangeMatchResult = -1
)

func fnMatchX(pattern, str string, patternOffset, strOffset int, flags FnMatchFlag, recurs int) (bool, error) {
	recurs--
	if recurs == 0 {
		return false, errors.New("too deep recursion")
	}
	initialStrOffset := strOffset
	for {
		if patternOffset == len(pattern) {
			if (flags&FNMLeadingDir != 0) && str[0] == '/' {
				return true, nil
			}
			return strOffset == len(str), nil
		}
		c := pattern[patternOffset]
		patternOffset++
		switch c {
		case '?':
			if strOffset == len(str) {
				return false, nil
			}
			s := str[strOffset]
			if s == '/' && (flags&FNMPathName != 0) {
				return false, nil
			}
			if s == '.' && (flags&FNMPeriod != 0) && ((initialStrOffset == strOffset) || ((flags&FNMPathName != 0) && (strOffset != 0) && (str[strOffset-1] == '/'))) {
				return false, nil
			}
			strOffset++
		case '*':
			if patternOffset < len(pattern) {
				c = pattern[patternOffset]
				if c == '*' {
					flags &= FNMPathName
					for c == '*' && patternOffset < len(pattern) {
						patternOffset++
						c = pattern[patternOffset]
					}
					if c == '/' {
						patternOffset++
					}
				}
			}
			s := str[strOffset]
			if s == '.' && (flags&FNMPeriod != 0) && ((initialStrOffset == strOffset) || ((flags&FNMPathName != 0) && (strOffset != 0) && (str[strOffset-1] == '/'))) {
				return false, nil
			}
			if patternOffset == len(pattern) {
				if flags&FNMPathName != 0 {
					return ((flags&FNMLeadingDir != 0) || strings.IndexByte(str[strOffset:], '/') == -1), nil
				} else {
					return true, nil
				}
			}
			c = pattern[patternOffset]
			if c == '/' && (flags&FNMPathName != 0) {
				i := strings.IndexByte(str[strOffset:], '/')
				if i == -1 {
					return false, nil
				}
				strOffset += i
				break
			}
			for strOffset < len(str) {
				test := str[strOffset]
				r, err := fnMatchX(pattern, str, patternOffset, strOffset, flags, recurs)
				if err != nil {
					return false, err
				}
				if err == nil && r {
					return true, nil
				}
				if test == '/' && (flags&FNMPathName != 0) {
					break
				}
				strOffset++
			}
			return false, nil
		case '[':
			if strOffset == len(str) {
				return false, nil
			}
			s := str[strOffset]
			if s == '/' && (flags&FNMPathName != 0) {
				return false, nil
			}
			if s == '.' && (flags&FNMPeriod != 0) && (initialStrOffset == strOffset) || ((flags&FNMPathName != 0) && (strOffset != 0) && (str[strOffset-1] == '/')) {
				return false, nil
			}
			switch rangeMatch(pattern, s, &patternOffset, flags) {
			case RangeMatch:
				break
			case RangeNoMatch:
				return false, nil
			case RangeError:
				if caseInsensitiveMatch(c, str, strOffset, flags) {
					strOffset++
				} else {
					return false, nil
				}
			}
		case '\\':
			if flags&FNMNoEscape == 0 {
				if patternOffset == len(pattern) {
					patternOffset--
					c = '\\'
				}
			}
			if caseInsensitiveMatch(c, str, strOffset, flags) {
				strOffset++
			} else {
				return false, nil
			}
		default:
			if caseInsensitiveMatch(c, str, strOffset, flags) {
				strOffset++
			} else {
				return false, nil
			}
		}
	}
}

func caseInsensitiveMatch(c byte, str string, strOffset int, flags FnMatchFlag) bool {
	if strOffset == len(str) {
		return false
	}
	s := str[strOffset]
	if c == s {
		return true
	}
	if flags&FNMCaseFold == 0 {
		return false
	}
	in := []byte{c, s}
	out := bytes.ToLower(in)
	return out[0] == out[1]
}

func toLower(c byte) byte {
	in := []byte{c}
	out := bytes.ToLower(in)
	return out[0]
}

func rangeMatch(pattern string, test byte, originalPatternOffset *int, flags FnMatchFlag) RangeMatchResult {
	patternOffset := *originalPatternOffset
	c := pattern[patternOffset]

	negate := (c == '!' || c == '^')
	if negate {
		patternOffset++
		if patternOffset == len(pattern) {
			return RangeError
		}
		c = pattern[patternOffset]
	}
	if flags&FNMCaseFold != 0 {
		test = toLower(test)
	}
	ok := false
	for {
		if c == '\\' && (flags&FNMNoEscape == 0) {
			patternOffset++
			if patternOffset == len(pattern) {
				return RangeError
			}
			c = pattern[patternOffset]
		}
		if c == '/' && (flags&FNMPathName != 0) {
			return RangeNoMatch
		}
		if flags&FNMCaseFold != 0 {
			c = toLower(c)
		}
		if c == '-' {
			if patternOffset == len(pattern)-2 {
				return RangeError
			}
			c2 := pattern[patternOffset+1]
			if c2 != ']' {
				patternOffset += 2
				if c2 == '\\' && (flags&FNMNoEscape == 0) {
					patternOffset++
					if patternOffset == len(pattern) {
						return RangeError
					}
					c2 = pattern[patternOffset]
				}
				if flags&FNMCaseFold != 0 {
					c2 = toLower(c2)
				}
				if c <= test && test <= c2 {
					ok = true
				}
			} else if c == test {
				ok = true
			}

		} else if c == test {
			ok = true
		}

		c = pattern[patternOffset]
		patternOffset++
		if c == ']' {
			break
		}
		if patternOffset == len(pattern) {
			return RangeError
		}
	}
	*originalPatternOffset = patternOffset
	if ok == negate {
		return RangeNoMatch
	} else {
		return RangeMatch
	}
}

func fnMatch(pattern, str string, flags FnMatchFlag) bool {
	result, _ := fnMatchX(pattern, str, 0, 0, flags, 64)
	return result
}
