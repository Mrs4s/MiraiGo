package client

import "fmt"

type Logger interface {
	Error(c *QQClient, msg string, args ...interface{})
	Warning(c *QQClient, msg string, args ...interface{})
	Info(c *QQClient, msg string, args ...interface{})
	Debug(c *QQClient, msg string, args ...interface{})
	Trace(c *QQClient, msg string, args ...interface{})
	Dump(c *QQClient, dump []byte, msg string, args ...interface{})
}

type eventLogger []func(*QQClient, *LogEvent)

func (l eventLogger) Error(c *QQClient, msg string, args ...interface{}) {
	l.dispatchLogEvent(c, &LogEvent{
		Type:    "ERROR",
		Message: fmt.Sprintf(msg, args...),
	})
}

func (l eventLogger) Warning(c *QQClient, msg string, args ...interface{}) {
	l.dispatchLogEvent(c, &LogEvent{
		Type:    "WARNING",
		Message: fmt.Sprintf(msg, args...),
	})
}

func (l eventLogger) Info(c *QQClient, msg string, args ...interface{}) {
	l.dispatchLogEvent(c, &LogEvent{
		Type:    "INFO",
		Message: fmt.Sprintf(msg, args...),
	})
}

func (l eventLogger) Debug(c *QQClient, msg string, args ...interface{}) {
	l.dispatchLogEvent(c, &LogEvent{
		Type:    "DEBUG",
		Message: fmt.Sprintf(msg, args...),
	})
}

func (l eventLogger) Trace(c *QQClient, msg string, args ...interface{}) {
	l.dispatchLogEvent(c, &LogEvent{
		Type:    "TRACE",
		Message: fmt.Sprintf(msg, args...),
	})
}

func (l eventLogger) Dump(c *QQClient, dump []byte, msg string, args ...interface{}) {
	l.dispatchLogEvent(c, &LogEvent{
		Type:    "DUMP",
		Message: fmt.Sprintf(msg, args...),
		Dump:    dump,
	})
}

func (l eventLogger) dispatchLogEvent(c *QQClient, e *LogEvent) {
	if l == nil {
		return
	}
	for _, f := range l {
		cover(func() {
			f(c, e)
		})
	}
}

// example below...
type nopLogger struct{}

func (l nopLogger) Error(c *QQClient, msg string, args ...interface{})   {}
func (l nopLogger) Warning(c *QQClient, msg string, args ...interface{}) {}
func (l nopLogger) Info(c *QQClient, msg string, args ...interface{})    {}
func (l nopLogger) Debug(c *QQClient, msg string, args ...interface{})   {}
func (l nopLogger) Trace(c *QQClient, msg string, args ...interface{})   {}
