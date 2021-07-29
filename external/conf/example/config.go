package main

type Config struct {
	Port        Port        `yaml:"port"`
	Env         string      `yaml:"env"`
	Apollo      Apollo      `yaml:"apollo"`
	Application Application `yaml:"application"`
}

type Port struct {
	GRPCAddr  string `yaml:"gRPCAddr"`
	AdminAddr string `yaml:"adminAddr"`
}

type Apollo struct {
	AppID     string `yaml:"appID"`
	Cluster   string `yaml:"cluster"`
	Namespace string `yaml:"namespace"`
	Meta      string `yaml:"meta"` //meta 替代 ip
}

type Application struct {
	Name string
}
