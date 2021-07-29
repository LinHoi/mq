package main

import (
	"conf"
	"fmt"
	"github.com/mohae/deepcopy"
	"os"
	"reflect"
	"time"
)

func main() {
	c := &Config{}
	directory,_ := os.Getwd()
	err := conf.New(c,conf.WithFile(directory+"/example/env.yaml"))
	if err != nil {
		fmt.Println("err",err)
	}
	fmt.Println(c)
	oldConfig := deepcopy.Copy(c)
	for {
		if !reflect.DeepEqual(oldConfig, c) {
			fmt.Printf("配置发生变化, 新配置：%v\n",c)
			oldConfig = deepcopy.Copy(c)
		}
		fmt.Println(c)
		time.Sleep(5*time.Second)
	}
}
