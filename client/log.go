package client

type Logger interface {
	Info(format string, args ...any)
	Warning(format string, args ...any)
	Error(format string, args ...any)
	Debug(format string, args ...any)
	Dump(dumped []byte, format string, args ...any)
}

func (c *QQClient) SetLogger(logger Logger) {
	c.logger = logger
}

func (c *QQClient) info(msg string, args ...any) {
	if c.logger != nil {
		c.logger.Info(msg, args...)
	}
}

func (c *QQClient) warning(msg string, args ...any) {
	if c.logger != nil {
		c.logger.Warning(msg, args...)
	}
}

func (c *QQClient) error(msg string, args ...any) {
	if c.logger != nil {
		c.logger.Error(msg, args...)
	}
}

func (c *QQClient) debug(msg string, args ...any) {
	if c.logger != nil {
		c.logger.Debug(msg, args...)
	}
}

func (c *QQClient) dump(msg string, data []byte, args ...any) {
	if c.logger != nil {
		c.logger.Dump(data, msg, args...)
	}
}
