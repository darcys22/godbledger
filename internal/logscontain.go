package internal

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	logTest "github.com/sirupsen/logrus/hooks/test"
)

type assertionLoggerFn func(string, ...interface{})

func parseMsg(defaultMsg string, msg ...interface{}) string {
	if len(msg) >= 1 {
		msgFormat, ok := msg[0].(string)
		if !ok {
			return defaultMsg
		}
		return fmt.Sprintf(msgFormat, msg[1:]...)
	}
	return defaultMsg
}

// LogsContain checks whether a given substring is a part of logs. If flag=false, inverse is checked.
func LogsContain(loggerFn assertionLoggerFn, hook *logTest.Hook, want string, flag bool, msg ...interface{}) {
	_, file, line, _ := runtime.Caller(2)
	entries := hook.AllEntries()
	var logs []string
	match := false
	for _, e := range entries {
		msg, err := e.String()
		if err != nil {
			loggerFn("%s:%d Failed to format log entry to string: %v", filepath.Base(file), line, err)
			return
		}
		if strings.Contains(msg, want) {
			match = true
		}
		for _, field := range e.Data {
			fieldStr, ok := field.(string)
			if !ok {
				continue
			}
			if strings.Contains(fieldStr, want) {
				match = true
			}
		}
		logs = append(logs, msg)
	}
	var errMsg string
	if flag && !match {
		errMsg = parseMsg("Expected log not found", msg...)
	} else if !flag && match {
		errMsg = parseMsg("Unexpected log found", msg...)
	}
	if errMsg != "" {
		loggerFn("%s:%d %s: %v\nSearched logs:\n%v", filepath.Base(file), line, errMsg, want, logs)
	}
}
