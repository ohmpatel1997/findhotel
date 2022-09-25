// Package zlog holds the logger and the log utility functions
package zlog

import (
	"fmt"
	"os"
	"sync"

	"github.com/rs/zerolog"
)

// ParamsType denotes the type of log functions extra parameters
type ParamsType map[string]interface{}

// Log represents zerolog logger
type Log struct {
	logger *zerolog.Logger
}

var once sync.Once
var (
	instance *Log
)

func New() *Log {
	once.Do(func() {
		var logger zerolog.Logger
		zerolog.LevelFieldName = "status"
		logger = zerolog.New(os.Stdout).With().Timestamp().Logger().Level(zerolog.InfoLevel)
		instance = &Log{
			logger: &logger,
		}
	})

	return instance
}

func Logger() *Log {
	if instance == nil {
		panic("Logger was not instantiated")
	}
	return instance
}

func (z *Log) Debug(msg string, fields map[string]interface{}) {
	if fields == nil {
		fields = make(map[string]interface{})
	}
	z.logger.Debug().Fields(fields).Msg(msg)
}

func (z *Log) Info(msg string, fields map[string]interface{}) {
	z.logger.Info().Fields(fields).Msg(msg)
}

func (z *Log) Warn(msg string, fields map[string]interface{}) {
	if fields == nil {
		fields = make(map[string]interface{})
	}
	z.logger.Warn().Fields(fields).Msg(msg)
}

func (z *Log) Error(msg string, err error, fields map[string]interface{}) {
	if fields == nil {
		fields = make(map[string]interface{})
	}

	if err != nil {
		fields["error.message"] = err.Error()
		msg = fmt.Sprintf("%s: %s", msg, err.Error())
	}

	z.logger.Error().Fields(fields).Msg(msg)
}

func (z *Log) PanicError(msg string, err error, stack string) {
	fields := map[string]interface{}{
		"error":       err.Error(),
		"error.stack": stack,
	}
	z.logger.Error().Fields(fields).Msg(msg)
}
