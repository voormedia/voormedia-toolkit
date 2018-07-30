package util

import (
	"io"
	"os"

	"github.com/fatih/color"
)

type Logger struct {
	prefix string
	out    io.Writer
	color  bool
}

func NewLogger(prefix string) *Logger {
	return &Logger{
		prefix: prefix,
		out:    os.Stderr,
		color:  true,
	}
}

func (log *Logger) Log(a ...interface{}) {
	col := color.New()
	col.Fprint(log.out, log.prefix+": ")
	col.Fprintln(log.out, a...)
}

func (log *Logger) Note(a ...interface{}) {
	col := color.New(color.Bold, color.FgCyan)
	col.Fprint(log.out, log.prefix+": ")
	col.Fprintln(log.out, a...)
}

func (log *Logger) Success(a ...interface{}) {
	col := color.New(color.Bold, color.FgGreen)
	col.Fprint(log.out, log.prefix+": ")
	col.Fprintln(log.out, a...)
}

func (log *Logger) Warn(a ...interface{}) {
	col := color.New(color.Bold, color.FgYellow)
	col.Fprint(log.out, log.prefix+": ")
	col.Fprintln(log.out, a...)
}

func (log *Logger) Error(a ...interface{}) {
	col := color.New(color.Bold, color.FgRed)
	col.Fprint(log.out, log.prefix+": ")
	col.Fprintln(log.out, a...)
}

func (log *Logger) Fatal(a ...interface{}) {
	log.Error(a...)
	os.Exit(1)
}
