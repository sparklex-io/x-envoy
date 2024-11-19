package log

import (
	"github.com/rs/zerolog"
	"os"
)

var lg zerolog.Logger

func init() {
	lg = zerolog.New(os.Stdout).Output(zerolog.NewConsoleWriter()).With().Caller().Timestamp().Logger()
}

func SetLoggerLevel(level zerolog.Level) {
	lg = lg.Level(level)
}

func Debug() *zerolog.Event {
	return lg.Debug()
}

func Info() *zerolog.Event {
	return lg.Info()
}

func Warn() *zerolog.Event {
	return lg.Warn()
}

func Error() *zerolog.Event {
	return lg.Error()
}
