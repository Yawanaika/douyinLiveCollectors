package log

import (
	"douyinLiveCollectors/backend/library/time"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"sync"
)

const (
	maxFileSize = 20 * 1024 * 1024 // 20MB
	logDir      = "./logs"
)

var (
	Logger   *DefaultLogger
	initOnce sync.Once
	mutex    sync.Mutex
)

type DefaultLogger struct {
	liveId    string
	file      *os.File
	size      int64
	logFolder string
	slogger   *slog.Logger
}

func NewLogger() *DefaultLogger {
	return &DefaultLogger{
		slogger: slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})),
	}
}

func (l *DefaultLogger) getLogFolderPath() (string, error) {
	date := time.Today()
	var dir string
	if l.liveId == "" {
		dir = filepath.Join(logDir, date)
	} else {
		dir = filepath.Join(logDir, date, l.liveId)
	}

	// 获取最新的文件夹序号
	files, err := os.ReadDir(dir)
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}

	var latestIndex int
	for _, file := range files {
		if file.IsDir() {
			index, err := strconv.Atoi(file.Name())
			if err == nil && index > latestIndex {
				latestIndex = index
			}
		}
	}

	latestIndex++
	subDir := filepath.Join(dir, strconv.Itoa(latestIndex))
	if err := os.MkdirAll(subDir, 0755); err != nil {
		return "", err
	}

	return subDir, nil
}

func (l *DefaultLogger) getLogFilePath() (string, error) {
	if l.logFolder == "" {
		var err error
		l.logFolder, err = l.getLogFolderPath()
		if err != nil {
			return "", err
		}
	}

	files, err := os.ReadDir(l.logFolder)
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}

	fileIndex := len(files) + 1
	var fileName string
	if l.liveId == "" {
		fileName = fmt.Sprintf("%d.log", fileIndex)
	} else {
		fileName = fmt.Sprintf("%s-%d.log", l.liveId, fileIndex)
	}
	return filepath.Join(l.logFolder, fileName), nil
}

func (l *DefaultLogger) openLogFile() error {
	path, err := l.getLogFilePath()
	if err != nil {
		return err
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	l.file = file
	l.size = getFileSize(file)
	l.slogger = slog.New(slog.NewTextHandler(file, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	return nil
}

func (l *DefaultLogger) rotateLogFile() error {
	if l.file != nil {
		l.file.Close()
		l.file = nil
	}
	return l.openLogFile()
}

func (l *DefaultLogger) writeLog(level slog.Level, message ...any) {
	mutex.Lock()
	defer mutex.Unlock()

	if l.file == nil {
		if err := l.openLogFile(); err != nil {
			l.slogger.Error("Failed to open log file", "error", err)
			return
		}
	}

	if l.size >= maxFileSize {
		if err := l.rotateLogFile(); err != nil {
			l.slogger.Error("Failed to rotate log file", "error", err)
			return
		}
	}

	//_, file, line, _ := runtime.Caller(2)
	logEntry := fmt.Sprintf("%s [%s] : %s\n", time.Now(), level, message)
	if _, err := l.file.WriteString(logEntry); err != nil {
		l.slogger.Error("Failed to write log entry", "error", err)
		return
	}

	l.size += int64(len(logEntry))
}

func getFileSize(file *os.File) int64 {
	info, _ := file.Stat()
	return info.Size()
}

func (l *DefaultLogger) SetLiveId(liveId string) {
	l.liveId = liveId
}

func GetLogger() *DefaultLogger {
	initOnce.Do(func() {
		Logger = NewLogger()
	})
	return Logger
}

func Debug(format string, a ...any) {
	Logger.writeLog(slog.LevelDebug, fmt.Sprintf(format, a...))
}

func Info(format string, a ...any) {
	Logger.writeLog(slog.LevelInfo, fmt.Sprintf(format, a...))
}

func Warn(format string, a ...any) {
	Logger.writeLog(slog.LevelWarn, fmt.Sprintf(format, a...))
}

func Error(format string, a ...any) {
	Logger.writeLog(slog.LevelError, fmt.Sprintf(format, a...))
}
