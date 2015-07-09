package git4go

import (
	"bytes"
	"errors"
	"strconv"
	"strings"
	"time"
	//"log"
)

func (repo *Repository) DefaultSignature() (*Signature, error) {
	config := repo.Config()
	var errorStrings []string
	name, err := config.LookupString("user.name")
	if err != nil {
		errorStrings = append(errorStrings, "can't get user.name")
	}
	email, err := config.LookupString("user.email")
	if err != nil {
		errorStrings = append(errorStrings, "can't get user.email")
	}
	if len(errorStrings) != 0 {
		return nil, errors.New(strings.Join(errorStrings, "\n"))
	}
	return &Signature{
		Name:  name,
		Email: email,
		When:  time.Now(),
	}, nil

}

type Signature struct {
	Name  string
	Email string
	When  time.Time
}

// the offset in mintes, which is what git wants
func (v *Signature) Offset() int {
	_, offset := v.When.Zone()
	return offset / 60
}

func parseSignature(data []byte, offset int, prefix []byte) (*Signature, int, error) {
	linePrefix := offset + len(prefix)
	lineEnd := offset
	found := false
	emailStart := -1
	emailEnd := -1
	for lineEnd < len(data) {
		switch data[lineEnd] {
		case '<':
			emailStart = lineEnd - linePrefix
		case '>':
			emailEnd = lineEnd - linePrefix
		case '\n':
			found = true
		}
		if found {
			break
		}
		lineEnd++
	}
	if !found {
		return nil, offset, errors.New("no newline given")
	}
	if !bytes.Equal(data[offset:offset+len(prefix)], prefix) {
		return nil, offset, errors.New("expected prefix doesn't match actual")
	}
	line := data[linePrefix:lineEnd]
	if emailStart == -1 || emailEnd == -1 || emailEnd < emailStart {
		return nil, offset, errors.New("malformed e-mail")
	}
	sig := &Signature{
		Name:  string(bytes.TrimSpace(line[:emailStart-1])),
		Email: string(bytes.TrimSpace(line[emailStart+1 : emailEnd])),
	}
	timeStart := emailEnd + 1
	for timeStart < len(line) && line[timeStart] == ' ' {
		timeStart++
	}
	timeEnd := timeStart + 1
	for timeEnd < len(line) && line[timeEnd] != ' ' {
		timeEnd++
	}
	epoch, err := strconv.ParseInt(string(line[timeStart:timeEnd]), 10, 64)
	if err != nil {
		return nil, offset, err
	}
	timestamp := time.Unix(epoch, 0)
	timezone, err := strconv.ParseInt(string(bytes.TrimSpace(line[timeEnd:])), 10, 64)
	hour := timezone / 100
	min := timezone % 100
	if hour < 14 && min < 59 {
		second := int(hour*3600 + min*60)
		_, localTimezone := timestamp.Zone()
		timestamp = timestamp.Add(time.Second * time.Duration(localTimezone-second))
		timestamp = timestamp.In(time.FixedZone(" ", second))
	}
	sig.When = timestamp
	return sig, lineEnd + 1, nil
}
