package log

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	rotatelogs "github.com/lestrrat/go-file-rotatelogs"
)

type Logger struct {
	Path      string `yaml:"path"`
	Duration  string `yaml:"duration"`
	DebugMode bool   `yaml:"debug"`
	Level     string `yaml:"level"`
}

var (
	_logging *log.Logger
	_logPath string
)

const (
	InfoPrefix  = "INFO "
	DebugPrefix = "DEBUG "
	ErrorPrefix = "ERROR "
	WarnPrefix  = "WARN "
	DebugMode   = iota
	ProductMode

	debugLevel = 1
	infoLevel  = 2
	warnLevel  = 3
	errorLevel = 4
)

var (
	levelMap = map[string]int{
		"info":  infoLevel,
		"debug": debugLevel,
		"warn":  warnLevel,
		"error": errorLevel,
	}

	myLevel = 0
)

func (l Logger) GetLogging() *log.Logger {
	return _logging
}

func (l Logger) StartLog() {
	level, ok := levelMap[l.Level]
	if !ok {
		level = 0
	}
	myLevel = level

	mode := DebugMode
	if !l.DebugMode {
		mode = ProductMode
	}
	var duration time.Duration
	switch l.Duration {
	case "day":
		duration = time.Hour * 24
	case "hour":
		duration = time.Hour
	case "minute":
		duration = time.Minute
	}

	_Init(l.Path, mode, duration)
}

func getPatternAndDuration(d time.Duration) (string, time.Duration) {
	switch d {
	case time.Hour:
		return "%Y%m%d%H", d
	case time.Minute:
		return "%Y%m%d%H%M", d
	default:
		return "%Y%m%d", time.Hour * 24
	}
}

func _Init(logPath string, mode int, rotationTime time.Duration) {
	var out io.Writer

	if mode == ProductMode {
		pattern, time := getPatternAndDuration(rotationTime)
		logFilePath := fmt.Sprintf("%s/%s.log", logPath, pattern)
		out, _ = rotatelogs.New(logFilePath, rotatelogs.WithRotationTime(time))
	} else {
		out = os.Stdout
	}
	_logging = log.New(out, InfoPrefix, log.Ldate|log.Lmicroseconds|log.Llongfile)
	_logPath = logPath
}

func (l Logger) Info(msg string) {
	if myLevel > infoLevel {
		return
	}
	_logging.SetPrefix(InfoPrefix)
	_logging.Output(2, msg)
}

func (l Logger) Debug(msg string) {
	if myLevel > debugLevel {
		return
	}
	_logging.SetPrefix(DebugPrefix)
	_logging.Output(2, msg)
}

func (l Logger) Warn(msg string) {
	if myLevel > warnLevel {
		return
	}
	_logging.SetPrefix(WarnPrefix)
	_logging.Output(2, msg)
}

func (l Logger) Err(msg string) {
	if myLevel > errorLevel {
		return
	}
	_logging.SetPrefix(ErrorPrefix)
	_logging.Output(2, msg)
}

func (l Logger) WriteFile(file string, data []byte) (string, error) {
	fullFilePath := fmt.Sprintf("%s/%s", _logPath, file)
	l.Info(fullFilePath)
	dir := filepath.Dir(fullFilePath)
	l.Info(dir)

	if exist, _ := pathExist(dir); !exist {
		err := os.MkdirAll(dir, 0744)
		if err != nil {
			l.Err(err.Error())
		} else {
			l.Warn(fmt.Sprintf("Create Dir: %s", dir))
		}
	}

	err := ioutil.WriteFile(fullFilePath, data, 744)
	if err != nil {
		l.Err(err.Error())
	}
	return fullFilePath, err
}

func pathExist(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func removeContents(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	return nil
}
