package git4go

// test from https://github.com/tarruda/node-git-core

import (
	"testing"
	"encoding/hex"
	"bytes"
	//"fmt"
)

func Test_Delta_and_Apply_1(t *testing.T) {
	a := []byte("text file line 1\ntext file line 2\na")
	b := []byte("text file line 2\ntext file line 1\nab")
	delta, err := CreateDelta(a, b, 0)
	if err != nil {
		t.Error("err should be nil:", err)
	} else {
		if hex.EncodeToString(delta) != "23249111119011026162" {
			t.Error("delta is wrong: ", hex.EncodeToString(delta))
		} else {
			c, err := ApplyDelta(a, delta)
			if err != nil {
				t.Error("err should be nil:", err)
			} else if bytes.Compare(b, c) != 0 {
				t.Error("patched data is wrong: ", string(b), string(c))
			}
		}
	}
}

var testSrc string = `some text
with some words
abcdef
ghijkl
mnopqr
ab
rst`

var testTarget string = `some text
words
abcdef
ghijkl
mnopqr
ba
rst
h`

func Test_Delta_and_Apply_2(t *testing.T) {
	a := []byte(testSrc)
	b := []byte(testTarget)
	delta, err := CreateDelta(a, b, 0)
	if err != nil {
		t.Error("err should be nil:", err)
	} else {
		if hex.EncodeToString(delta) != "352d900b056f7264730a911a150862610a7273740a68" {
			t.Error("delta is wrong: ", hex.EncodeToString(delta))
		} else {
			c, err := ApplyDelta(a, delta)
			if err != nil {
				t.Error("err should be nil:", err)
			} else if bytes.Compare(b, c) != 0 {
				t.Error("patched data is wrong: ", string(b), string(c))
			}
		}
	}
}

func Test_Delta_and_Apply_BinaryData(t *testing.T) {
	var a bytes.Buffer
	for i := 0; i < (1 << 14); i++ {
		a.WriteByte(200)
	}

	var b bytes.Buffer
	for i := 0; i < (1 << 13) - 10; i++ {
		a.WriteByte(200)
	}
	for i := 0; i < 10; i++ {
		a.WriteByte(199)
	}
	for i := 0; i < (1 << 13); i++ {
		a.WriteByte(200)
	}

	delta, err := CreateDelta(a.Bytes(), b.Bytes(), 0)
	if err != nil {
		t.Error("err should be nil:", err)
	} else {
		c, err := ApplyDelta(a.Bytes(), delta)
		if err != nil {
			t.Error("err should be nil:", err)
		} else if bytes.Compare(b.Bytes(), c) != 0 {
			t.Error("patched data is wrong")
		}
	}
}
