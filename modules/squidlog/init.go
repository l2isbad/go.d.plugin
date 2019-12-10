package squidlog

import (
	"bytes"
	"fmt"

	"github.com/netdata/go.d.plugin/pkg/logs"
)

func (s *SquidLog) createLogLine() {
	s.line = &logLine{}
}

func (s *SquidLog) createLogReader() error {
	s.Cleanup()
	s.Debug("starting log reader creating")
	reader, err := logs.Open(s.Path, s.ExcludePath, s.Logger)
	if err != nil {
		return fmt.Errorf("creating log reader: %v", err)
	}
	s.Debugf("created log reader, current file '%s'", reader.CurrentFilename())
	s.file = reader
	return nil
}

func (s *SquidLog) createParser() error {
	s.Debug("starting parser creating")
	lastLine, err := logs.ReadLastLine(s.file.CurrentFilename(), 0)
	if err != nil {
		return fmt.Errorf("read last line: %v", err)
	}
	lastLine = bytes.TrimRight(lastLine, "\n")
	s.Debugf("last line: '%s'", string(lastLine))

	s.parser, err = logs.NewParser(s.Parser, s.file)
	if err != nil {
		return fmt.Errorf("create parser: %v", err)
	}
	s.Debugf("created parser: %s", s.parser.Info())

	err = s.parser.Parse(lastLine, s.line)
	if err != nil {
		return fmt.Errorf("parse last line: %v (%s)", err, string(lastLine))
	}

	if err = s.line.verify(); err != nil {
		return fmt.Errorf("verify last line: %v (%s)", err, string(lastLine))
	}
	return nil
}