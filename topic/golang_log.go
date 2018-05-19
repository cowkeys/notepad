package log

import (
	"time"

    "os"
	"gitlab.followme.com/FollowmeGo/utils/env"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var LEVELS = map[string]zapcore.Level{
	"debug": zap.DebugLevel,
	"info":  zap.InfoLevel,
	"warn":  zap.WarnLevel,
	"err":   zap.ErrorLevel,
}

type Logger struct {
	path  string
	rlog  *lumberjack.Logger
	log   *zap.Logger
	sugar *zap.SugaredLogger

	level zapcore.Level
	pid   []interface{}

	rolling    bool
	lastRotate time.Time
}

func NewLogger(path string, level string) (*Logger, error) {
	out := new(Logger)
	out.rlog = new(lumberjack.Logger)

	out.path = path
	out.lastRotate = time.Now()
	out.level = LEVELS[level]
	out.pid = []interface{}{os.Getpid()}

	// config lumberjack
	out.rlog.Filename = path
	out.rlog.MaxSize = 0x1000 * 2 // automatic rolling file on it increment than 2GB
	out.rlog.LocalTime = true
	out.rlog.Compress = true
	out.rlog.MaxBackups = 60 // reserve last 60 day logs

	// config encoder config
	ec := zap.NewProductionEncoderConfig()
	ec.EncodeLevel = zapcore.CapitalLevelEncoder
	ec.EncodeTime = zapcore.ISO8601TimeEncoder

	// config core
	c := zapcore.AddSync(out.rlog)
	core := zapcore.NewCore(zapcore.NewJSONEncoder(ec), c, out.level)
	out.log = zap.New(
		core,
		zap.AddCaller(),
		zap.AddCallerSkip(2),
	).
		With(zap.Int("pid", os.Getpid()))

	out.sugar = out.log.Sugar()
	return out, nil
}

func (tlog *Logger) checkRotate() {
	if !tlog.rolling {
		return
	}

	n := time.Now()
	y, m, d := tlog.lastRotate.Year(), tlog.lastRotate.Month(), tlog.lastRotate.Day()
	if y != n.Year() || m != n.Month() || d != n.Day() {
		tlog.rlog.Rotate()
		tlog.lastRotate = n
	}
}

func (tlog *Logger) EnableDailyFile() {
	tlog.rolling = true
}

func (tlog *Logger) Err(format string, v ...interface{}) {
	tlog.checkRotate()
	if !tlog.level.Enabled(zap.ErrorLevel) {
		return
	}

	defer tlog.log.Sync()
	tlog.sugar.Errorf(format, v...)
}

func (tlog *Logger) Errw(format string, v ...interface{}) {
	tlog.checkRotate()
	if !tlog.level.Enabled(zap.ErrorLevel) {
		return
	}

	defer tlog.log.Sync()
	tlog.sugar.Errorw(format, v...)
}

func (tlog *Logger) Warn(format string, v ...interface{}) {
	tlog.checkRotate()
	if !tlog.level.Enabled(zap.WarnLevel) {
		return
	}

	defer tlog.log.Sync()
	tlog.sugar.Warnf(format, v...)
}

func (tlog *Logger) Warnw(format string, v ...interface{}) {
	tlog.checkRotate()
	if !tlog.level.Enabled(zap.WarnLevel) {
		return
	}

	defer tlog.log.Sync()
	tlog.sugar.Warnw(format, v...)
}

func (tlog *Logger) Info(format string, v ...interface{}) {
	tlog.checkRotate()
	if !tlog.level.Enabled(zap.InfoLevel) {
		return
	}

	defer tlog.log.Sync()
	tlog.sugar.Infof(format, v...)
}

func (tlog *Logger) Infow(format string, v ...interface{}) {
	tlog.checkRotate()
	if !tlog.level.Enabled(zap.InfoLevel) {
		return
	}

	defer tlog.log.Sync()
	tlog.sugar.Infow(format, v...)
}

func (tlog *Logger) Debug(format string, v ...interface{}) {
	tlog.checkRotate()
	if !tlog.level.Enabled(zap.DebugLevel) {
		return
	}

	defer tlog.log.Sync()
	tlog.sugar.Debugf(format, v...)
}

func (tlog *Logger) Debugw(format string, v ...interface{}) {
	tlog.checkRotate()
	if !tlog.level.Enabled(zap.DebugLevel) {
		return
	}

	defer tlog.log.Sync()
	tlog.sugar.Debugw(format, v...)
}

var _logger *Logger

func GetDefault() *Logger {
	return _logger
}

func SetDefault(l *Logger) {
	_logger = l
}

func Stdout() {
	l, _ := NewLogger("stdout", "debug")
	SetDefault(l)
}

// deleted

func Emerg(format string, v ...interface{}) {
	_logger.Err(format, v...)
}

func Alert(format string, v ...interface{}) {
	_logger.Err(format, v...)
}

func Crit(format string, v ...interface{}) {
	_logger.Err(format, v...)
}

func Notice(format string, v ...interface{}) {
	_logger.Info(format, v...)
}

// active

func Err(format string, v ...interface{}) {
	_logger.Err(format, v...)
}

func Warn(format string, v ...interface{}) {
	_logger.Warn(format, v...)
}

func Info(format string, v ...interface{}) {
	_logger.Info(format, v...)
}

func Debug(format string, v ...interface{}) {
	_logger.Debug(format, v...)
}

func Errw(format string, v ...interface{}) {
	_logger.Errw(format, v...)
}

func Warnw(format string, v ...interface{}) {
	_logger.Warnw(format, v...)
}

func Infow(format string, v ...interface{}) {
	_logger.Infow(format, v...)
}

func Debugw(format string, v ...interface{}) {
	_logger.Debugw(format, v...)
}
