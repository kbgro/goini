package goini

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const (
	kv  Token = 1
	sec Token = 2
)

type Token int

type Section map[string]interface{}

type IniPayload map[string]interface{}

type Ini struct {
	path      string
	file      *os.File
	currToken Token
}

func NewIni(path string) *Ini {
	return &Ini{
		path:      path,
		file:      nil,
		currToken: 1,
	}
}

func (i *Ini) Open() error {
	f, err := os.Open(i.path)
	if err != nil {
		return err
	}
	i.file = f
	return nil
}

func (i *Ini) Close() error {
	if i.file != nil {
		return i.file.Close()
	}
	return nil
}

func (i *Ini) Parse() (IniPayload, error) {
	tempSec := ""
	payload := make(IniPayload)
	scanner := bufio.NewScanner(i.file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) == 0 || i.isComment(line) {
			continue
		}
		isSec, err := i.isSection(line)
		if err != nil {
			return nil, err
		}
		if isSec {
			tempSec, err = i.sectionHeader(line)
			if err != nil {
				return nil, err
			}
			payload[tempSec] = make(Section)
			i.currToken = sec
		}

		k, v, err := i.tokens(line)
		if i.currToken == kv {
			payload[k] = v
		} else {
			payload[tempSec].(Section)[k] = v
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return payload, nil
}

func (i *Ini) isComment(line string) bool {
	return line[0] == ';'
}

func (i *Ini) sectionHeader(line string) (string, error) {
	r := regexp.MustCompile(`^\[(\S+)\]$`)
	s := r.FindStringSubmatch(line)
	if len(s) == 2 {
		return s[1], nil
	}
	return "", errors.New("invalid section")
}

func (i *Ini) isSection(line string) (bool, error) {
	return regexp.Match(`^\[(\S+)\]$`, []byte(line))
}

func (i *Ini) tokens(line string) (string, interface{}, error) {
	result := strings.SplitN(line, "=", 2)
	if len(result) != 2 {
		return "", "", errors.New(fmt.Sprintf("invalid format %s", line))
	}
	var v interface{}
	k, v := strings.TrimSpace(result[0]), strings.TrimSpace(result[1])
	if ok, err := regexp.Match(`^\d+$`, []byte(v.(string))); ok && err == nil {
		v, err = strconv.Atoi(v.(string))
	}
	return k, v, nil
}
