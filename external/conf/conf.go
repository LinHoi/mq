package conf

import (
	"github.com/pkg/errors"
	"reflect"
)

type Option func(*options)

type options struct {
	file   string
	apollo *Apollo // apollo配置信息
	apolloSwitch bool
}

func WithFile(file string) Option {
	return func(o *options) {
		o.file = file
	}
}

func checkFile() Option {
	return func(o *options) {
		if o.file == "" {
			panic("file name is required")
		}
	}
}

func WithApollo(apollo *Apollo) Option {
	return func(o *options) {
		o.apollo = apollo
	}
}

func WithoutApollo() Option {
	return func(o *options) {
		o.apolloSwitch = false
	}
}

// New create a new config
func New(config interface{}, opts ...Option) error {
	if reflect.ValueOf(config).Type().Kind() != reflect.Ptr {
		return errors.New("pointer is required")
	}
	o := newOption(opts...)
	manager := &Manager{
		file:   o.file,
		config: config,
		apollo: o.apollo,
	}

	err := manager.getConfFromFile()
	if err != nil {
		return nil
	}

	if o.apolloSwitch == false {
		return nil
	}

	err = manager.getConfFromApollo()
	return err
}

func newOption(opts ...Option) *options {
	o := &options{}
	opts = append(opts, checkFile())
	for _, opt := range opts {
		opt(o)
	}

	return o
}
