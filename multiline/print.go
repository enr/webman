package multiline

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/candrewlee14/webman/ui"
	"github.com/fatih/color"
)

const esc = "\033["

var (
	ClearLine  = code(esc + "2K\r")
	MoveUp     = code(esc + "1F")
	MoveDown   = code(esc + "1E")
	ShowCursor = code(esc + "?25h")
	HideCursor = code(esc + "?25l")
)

func code(s string) []byte {
	if ui.AreAnsiCodesEnabled() {
		return []byte(s)
	}
	return []byte{}
}

type LineLogger struct {
	index  int
	count  int
	prefix string
	w      io.Writer
}

func (l *LineLogger) SetPrefix(pref string) {
	l.prefix = pref
}

func (l *LineLogger) Printf(format string, a ...any) {
	if len(MoveUp) != 0 {
		for i := 0; i < l.count-l.index; i++ {
			fmt.Fprintf(l.w, "%s", MoveUp)
		}
		fmt.Fprintf(l.w, "%s", ClearLine)
	}
	fmt.Fprintf(l.w, l.prefix+format, a...)
	for i := 0; i < l.count-l.index; i++ {
		if len(MoveDown) == 0 {
			fmt.Fprintf(l.w, "\n")
			break
		}
		fmt.Fprintf(l.w, "%s", MoveDown)
	}
}

type MultiLogger struct {
	mu      sync.Mutex
	loggers []LineLogger
}

func New(count int, w io.Writer) MultiLogger {
	for i := 0; i < count; i++ {
		fmt.Fprintf(w, "\n")
	}
	loggers := make([]LineLogger, count)
	for i := 0; i < count; i++ {
		loggers[i] = LineLogger{index: i, count: count, prefix: "", w: w}
	}
	return MultiLogger{
		loggers: loggers,
	}
}

func (ml *MultiLogger) Printf(index int, format string, a ...any) {
	ml.mu.Lock()
	defer ml.mu.Unlock()
	ml.loggers[index].Printf(format, a...)
}

func (ml *MultiLogger) SetPrefix(index int, pref string) {
	ml.loggers[index].SetPrefix(pref)
}

func (ml *MultiLogger) PrintUntilDone(index int, printStr string, done <-chan bool, millis int) {
	go func() {
		i := 0
		for {
			select {
			case <-done:
				if !ui.AreAnsiCodesEnabled() {
					ml.Printf(index, printStr)
				}
				return
			default:
				if ui.AreAnsiCodesEnabled() {
					ml.Printf(index, printStr+" "+color.HiBlackString(strings.Repeat(".", i)))
				}
			}
			time.Sleep(time.Duration(millis) * time.Millisecond)
			i += 1
		}
	}()
}
