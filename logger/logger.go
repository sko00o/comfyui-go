package logger

import (
	"fmt"
	"log"
	"strings"
)

// Logger is the interface for logging messages.
type Logger interface {
	Debugf(tmpl string, args ...any)
	Infof(tmpl string, args ...any)
	Warnf(tmpl string, args ...any)
	Errorf(tmpl string, args ...any)
}

type LoggerExtend interface {
	Logger
	With(kv ...any) LoggerExtend
}

func NewStd() *StdLogger {
	return &StdLogger{}
}

// StdLogger uses log package to implement Logger interface
type StdLogger struct {
	kv     []string
	prefix string
}

func (l *StdLogger) Debugf(tmpl string, args ...any) {
	log.Printf("[DBG] "+l.prefix+" "+tmpl, args...)
}

func (l *StdLogger) Infof(tmpl string, args ...any) {
	log.Printf("[INF] "+l.prefix+" "+tmpl, args...)
}

func (l *StdLogger) Warnf(tmpl string, args ...any) {
	log.Printf("[WRN] "+l.prefix+" "+tmpl, args...)
}

func (l *StdLogger) Errorf(tmpl string, args ...any) {
	log.Printf("[ERR] "+l.prefix+" "+tmpl, args...)
}

func (l *StdLogger) With(kv ...any) LoggerExtend {
	copy := *l

	for i := 0; i < len(kv); i += 2 {
		copy.kv = append(copy.kv, fmt.Sprintf("%s=%v", kv[i], kv[i+1]))
	}
	copy.prefix = strings.Join(copy.kv, " ")
	return &copy
}

var _ Logger = (*StdLogger)(nil)
