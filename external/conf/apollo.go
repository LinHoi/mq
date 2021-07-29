package conf

import "reflect"

type Apollo struct {
	AppID string `yaml:"appID"`
	Meta  string `yaml:"meta"`
}

func (a *Apollo) resolve(config interface{}) {
	var t reflect.Type
	var v reflect.Value
	if reflect.ValueOf(config).Type().Kind() == reflect.Ptr {
		t = reflect.TypeOf(config).Elem()
		v = reflect.ValueOf(config).Elem()
	} else {
		t = reflect.TypeOf(config)
		v = reflect.ValueOf(config)
	}

	apollo := resolveInterface(t, v, "Apollo")
	if apollo == nil {
		return
	}

	t = reflect.TypeOf(apollo)
	v = reflect.ValueOf(apollo)
	a.AppID = resolveString(t, v, "AppID")
	a.Meta = resolveString(t, v, "Meta")
}

func resolveInterface(t reflect.Type, v reflect.Value, key string) interface{} {
	if _, exist := t.FieldByName(key); exist {
		return v.FieldByName(key).Interface()
	}
	return nil
}

func resolveString(t reflect.Type, v reflect.Value, key string) string {
	if _, exist := t.FieldByName(key); exist {
		return v.FieldByName(key).String()
	}
	return ""
}