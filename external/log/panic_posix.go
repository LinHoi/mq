// +build !windows

package log

import (
	"os"
	"path/filepath"
	"syscall"

	"github.com/pkg/errors"
)

const panicFileName = "/panic.log"

// 系统发生panic即会输出到panic文件（recover的panic也会输出）
// windows系统暂时只会创建panic
func recordPanic(logPath string) error {
	if logPath == "" {
		return nil
	}

	err := os.MkdirAll(filepath.Dir(logPath), 0644)
	if err != nil {
		return errors.Wrap(err, "创建目录失败")
	}

	f, err := os.OpenFile(filepath.Join(filepath.Dir(logPath), panicFileName), os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		return errors.Wrap(err, "创建panic.log失败")
	}

	return errors.WithStack(syscall.Dup2(int(f.Fd()), int(os.Stderr.Fd())))
}
