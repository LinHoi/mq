package conf

import (
	"github.com/fsnotify/fsnotify"
	"github.com/mohae/deepcopy"
	"github.com/pkg/errors"
	"github.com/shima-park/agollo"
	remote "github.com/shima-park/agollo/viper-remote"
	"github.com/spf13/viper"
	"reflect"
	"strings"
	"time"
)

type provider string

const (
	confType                         = "yaml"
	apolloName                       = "apollo"
	dynamicConfigRefleshTimeInterval = time.Second

	fileProvider   provider = "file"
	apolloProvider provider = "apollo"
)

type Manager struct {
	file   string      //文件名称
	config interface{} // 全局配置的指针
	apollo *Apollo     // apollo配置信息
}

func (m Manager) getConfFromFile() error {
	v, err := m.newViper(fileProvider)
	if err != nil {
		return err
	}

	err = v.ReadInConfig()
	if err != nil {
		return err
	}

	conf := m.config
	err = v.Unmarshal(conf)
	if err == nil {
		m.config = conf
	}

	return err
}

// 目前只支持 file 或者 apollo provider
func (m Manager) newViper(pv provider) (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigType(confType)

	switch pv {
	case fileProvider:
		v.SetConfigFile(m.file)
		v.WatchConfig()
		v.OnConfigChange(func(in fsnotify.Event) {
			m.onConfigChange(v, pv)
		})

	case apolloProvider:
		apollo := m.getApollo()
		if len(apollo.AppID) == 0 || len(apollo.Meta) == 0 {
			return nil, errors.New("apollo config is required")
		}
		remote.SetAppID(apollo.AppID)
		remote.SetAgolloOptions(agollo.AutoFetchOnCacheMiss(), agollo.EnableSLB(true))
		err := v.AddRemoteProvider(apolloName, apollo.Meta, m.getFileName())
		if err != nil {
			return nil, err
		}

		err = v.WatchRemoteConfigOnChannel()
		if err != nil {
			return nil, err
		}
		go func() {
			for {
				time.Sleep(dynamicConfigRefleshTimeInterval)
				m.onConfigChange(v, pv)
			}
		}()
	default:
		return nil, errors.New("un support provider")
	}

	return v, nil
}

func (m Manager) updateConfig(newConfig interface{}) {
	if reflect.TypeOf(m.config).Kind() == reflect.Ptr {
		if reflect.ValueOf(m.config).Elem().CanSet() {
			reflect.ValueOf(m.config).Elem().Set(reflect.ValueOf(newConfig).Elem())
		}
	}
}

func (m Manager) getConfFromApollo() error {
	v, err := m.newViper(apolloProvider)
	if err != nil {
		return err
	}

	err = v.ReadRemoteConfig()
	if err != nil {
		return err
	}

	config := m.config
	err = v.Unmarshal(config)
	if err != nil {
		return err
	}

	m.config = config
	return nil
}

func (m *Manager) getApollo() *Apollo {
	if m.apollo != nil {
		return m.apollo
	}

	config := m.config
	apollo := &Apollo{}
	apollo.resolve(config)
	m.apollo = apollo
	return m.apollo
}

func (m *Manager) getFileName() string {
	paths := strings.Split(m.file, "/")
	if len(paths) > 1 {
		return paths[len(paths)-1]
	}
	return ""
}

func (m Manager) onConfigChange(v *viper.Viper, pv provider) {
	if v == nil {
		return
	}

	currentConfig := m.config
	dynamicConfig := deepcopy.Copy(currentConfig)
	err := v.Unmarshal(dynamicConfig)
	if err != nil {
		return
	}

	if !reflect.DeepEqual(currentConfig, dynamicConfig) {
		logger.Infow("update config from", "source", pv)
		m.updateConfig(dynamicConfig)
	}

}
