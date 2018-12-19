package weblog

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type groupMap map[string]string

func (gm groupMap) has(key string) bool {
	_, ok := gm[key]
	return ok
}

func (gm groupMap) get(key string) string {
	return gm[key]
}

func (gm groupMap) lookup(key string) (string, bool) {
	v, ok := gm[key]
	return v, ok
}

func newCSVParser(pattern csvPattern) *csvParser {
	return &csvParser{
		pattern: pattern,
		reader: csvReader{
			comma: ' ',
		},
		data: make(groupMap),
	}
}

type (
	parser interface {
		parse(line string) (groupMap, bool)
		info() string
	}

	csvParser struct {
		pattern csvPattern
		reader  csvReader

		data groupMap
	}
)

func (cp csvParser) info() string {
	var info []string

	for _, v := range cp.pattern {
		info = append(info, v.Name)
	}

	return fmt.Sprintf("[%s]", strings.Join(info, ", "))
}

func (cp *csvParser) parse(line string) (groupMap, bool) {
	lines, err := cp.reader.readRecord(line)

	if err != nil {
		return nil, false
	}

	if cp.pattern.max() > len(lines) {
		return nil, false
	}

	for _, f := range cp.pattern {
		cp.data[f.Name] = lines[f.Index]
	}

	return cp.data, true
}

func newParser(line string, patterns ...csvPattern) (parser, error) {
	if line == "" {
		return nil, errors.New("empty line")
	}

	for _, pattern := range patterns {
		if !pattern.isSorted() {
			return nil, fmt.Errorf("pattern %v is not sorted", pattern)
		}

		parser := newCSVParser(pattern)

		gm, ok := parser.parse(line)
		if !ok {
			continue
		}

		if err := validateResult(gm); err != nil {
			return nil, err
		}

		return parser, nil
	}

	return nil, errors.New("can't find appropriate csv parser")
}

func validateResult(gm map[string]string) error {
	_, ok := gm[keyCode]
	if !ok {
		return errors.New("mandatory key 'code' is missing")
	}

	for k, v := range gm {
		switch k {
		case keyCode:
			if !reCode.MatchString(v) {
				return fmt.Errorf("'code' bad syntax: '%s'", v)
			}
		case keyAddress:
			if !reAddress.MatchString(v) {
				return fmt.Errorf("'address' bad syntax: '%s'", v)
			}
		case keyBytesSent:
			if !reBytesSent.MatchString(v) {
				return fmt.Errorf("'bytes_sent' bad syntax: '%s'", v)
			}
		case keyRespLength:
			if !reResponseLength.MatchString(v) {
				return fmt.Errorf("'response_length' bad syntax: '%s'", v)
			}
		case keyRespTime, keyRespTimeUpstream:
			if !reResponseTime.MatchString(v) {
				return fmt.Errorf("'response_time' bad syntax : '%s'", v)
			}
		case keyRequest:
			gm, ok := reqParser.parse(v)
			if !ok {
				return fmt.Errorf("unparsable 'request' field : '%s'", v)
			}
			if !reHTTPMethod.MatchString(gm.get(keyMethod)) {
				return fmt.Errorf("'http_method' bad syntax : '%s'", v)
			}
			if !reHTTPVersion.MatchString(gm.get(keyVersion)) {
				return fmt.Errorf("'http_version' bad syntax : '%s'", v)
			}
		}
	}

	return nil
}

var (
	reAddress        = regexp.MustCompile(`[\da-f.:]+|localhost`)
	reCode           = regexp.MustCompile(`[1-9]\d{2}`)
	reBytesSent      = regexp.MustCompile(`\d+|-`)
	reResponseLength = regexp.MustCompile(`\d+|-`)
	reResponseTime   = regexp.MustCompile(`\d+|\d+\.\d+|-`)
	reHTTPMethod     = regexp.MustCompile(`[A-Z]+`)
	reHTTPVersion    = regexp.MustCompile(`HTTP/[0-9.]+`)
)

var reqParser = newCSVParser(csvPattern{
	{"method", 0},
	{"url", 1},
	{"version", 2},
})
