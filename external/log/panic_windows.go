// +build windows

package log

// windows系统暂时不处理
func recordPanic(logPath string) error {
	return nil
}
