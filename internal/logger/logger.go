package logger

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type Level int

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

type logEntry struct {
	level  Level
	format string
	args   []any
	time   time.Time
}

var (
	level             = InfoLevel
	output  io.Writer = os.Stdout
	logChan chan logEntry
	once    sync.Once
	wg      sync.WaitGroup
)

func Init(debug bool, outputs ...io.Writer) {
	once.Do(func() {
		if debug {
			level = DebugLevel
		} else {
			level = InfoLevel
		}

		if len(outputs) == 0 {
			output = os.Stdout
		} else if len(outputs) == 1 {
			output = outputs[0]
		} else {
			output = io.MultiWriter(outputs...)
		}

		logChan = make(chan logEntry, 128)
		wg.Add(1)
		go writer()
	})
}

func writer() {
	defer wg.Done()
	for entry := range logChan {
		if entry.level < level {
			continue
		}
		prefix := ""
		switch entry.level {
		case DebugLevel:
			prefix = "DEBUG"
		case InfoLevel:
			prefix = "INFO"
		case WarnLevel:
			prefix = "WARN"
		case ErrorLevel:
			prefix = "ERROR"
		case FatalLevel:
			prefix = "FATAL"
		}
		_, err := fmt.Fprintf(output, "%s [%s]: %s\n", entry.time.Format("02.01.2006 15:04:05"), prefix, fmt.Sprintf(entry.format, entry.args...))
		if err != nil {
			log.Fatal("could not write to output:", err)
		}
	}
}

func Close() {
	if logChan != nil {
		close(logChan)
		wg.Wait()
	}
}

func logf(lvl Level, format string, args ...any) {
	if logChan == nil {
		Init(false)
	}
	logChan <- logEntry{
		level:  lvl,
		format: format,
		args:   args,
		time:   time.Now(),
	}
}

func Debug(format string, args ...any) { logf(DebugLevel, format, args...) }
func Info(format string, args ...any)  { logf(InfoLevel, format, args...) }
func Warn(format string, args ...any)  { logf(WarnLevel, format, args...) }
func Error(format string, args ...any) { logf(ErrorLevel, format, args...) }

func Request(r *http.Request) {
	Debug("(%s) %s %s", r.RemoteAddr, r.Method, r.URL.String())
}

func InvalidMethod(r *http.Request) {
	Error("method not allowed for %s: %v", r.URL.String(), r.Method)
}
func Fatal(format string, args ...any) {
	logf(FatalLevel, format, args...)
	Close()
	os.Exit(1)
}
