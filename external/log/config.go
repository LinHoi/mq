// Copyright 2019 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package log

import "go.uber.org/zap/zapcore"

// Config serializes log related config in toml/json.
type Config struct {
	ServiceName string `yaml:"serviceName"`
	// 日志级别.
	Level string `yaml:"level"`
	// Log format. one of json or plain.
	Format string `yaml:"format"`
	// 是否输出到控制台.
	Stdout bool `yaml:"stdout"`
	// File log config.
	File FileLogConfig `yaml:"file"`
}

// level.
func (c *Config) level() zapcore.Level {
	level := zapcore.InfoLevel
	if c.Level == "" {
		return level
	} else {
		if err := level.Set(c.Level); err != nil {
			panic(err)
		}
		return level
	}
}

// FileLogConfig serializes file log related config.
type FileLogConfig struct {
	// 日志文件路径.
	Filename string `yaml:"filename"`
	// Is log rotate enabled.
	LogRotate bool `yaml:"logRotate"`
	// Max size for a single file, in MB.
	MaxSize int `yaml:"maxSize"`
	// Max log keep days, default is never deleting.
	MaxDays int `yaml:"maxDays"`
	// Maximum number of old log files to retain.
	MaxBackups int `yaml:"maxBackups"`
	// MAX size of bufio.Writer
	BufSize int `yaml:"bufSize"`
}
