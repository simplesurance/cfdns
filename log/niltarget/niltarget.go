// Package niltarget is a log driver that does nothing, dropping all messages.
package niltarget

import (
	"github.com/simplesurance/cfdns/log"
)

func New() log.Driver {
	return &logger{}
}

type logger struct{}

func (l *logger) Send(entry *log.Entry) {}
func (l *logger) PreLog() func()        { return nil }
