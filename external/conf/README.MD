# go 项目配置文件管理

配置是一个项目的基本组件，配置的组织方式会作用于项目的维护。糟糕的组织方式让项目的维护变得困难，曾经的项目中有通过机器环境变量来作为配置的，有通过机器本地的配置文件来加载配置的，有通过apollo 来实时获取配置的，也有通过flag 来传入配置参数的，配置的格式也是各式各样：普通的字符串，json，yaml 各种配置格式群魔乱舞。这些配置没有形成一个统一的入口，在代码编写上各自维护一套，使得配置这一基础组件无法收拢，不仅不利于项目的维护和交接，也无法在开发中形成一致的开发体验。

# 1. 对配置的基本述求

对于项目配置，开发者的最基本诉求是什么？在我个人的开发历程中，基本诉求是这样的：

- 1. 统一的配置文件管理入口

     配置文件统一管理，能够进行动态更新，也能配合服务的回滚实时回滚配置,支持分布式。

- 2. 统一的配置获取入口

     通过一个 Config  的 struct 就可以获取全部配置，类似代码如下

    ```go
     type Config struct {
       Application    Application     
       Port           Port            
       Database       Database                   
       Trace          TraceConf  
       Logger         LoggerConf                                 
       MQ             MQConf    
       Deamon         DeamonConf         
       Job            JobConf   
       //...
     }
    ```

- 3. 动态获取配置

     只要引用的是 Config 结构体的指针，取得的值都是Config 的实时配置

- 4. 所有的动态配置通过 一个协程来更新

     更新的时候只会更新有变化且不为空的配置，没有变化的配置保持不变，



# 2. 基本述求实现调研

## 2.1 配置中心选型

如何实现基本述求，最需要回答的问题是如何统一的配置文件管理入口。

传统的配置管理入口有机器本地的配置文件和环境变量。但这些配置扩展性不强，都不能进行及时的回滚和配置更新，对于分布式的支持也不友好。环境变量只能存储一些简单的键值对。机器本地配置文件需要在机器上更改，根据分工的不同，这一动作可能由SRE负责，会增加隐形沟通成本。

基于本地配置的配置管理方式是不好维护的。需要改变思路探索一种远程配置更新方式。这种远程配置应该有统一的配置管理入口，实时更新且能实时回滚，且应该是分布式的。



这里我的选择是 [apollo](https://www.apolloconfig.com/#/zh/README)

Apollo（阿波罗）是携程框架部门研发的分布式配置中心，能够集中化管理应用不同环境、不同集群的配置，配置修改后能够实时推送到应用端，并且具备规范的权限、流程治理等特性，适用于微服务配置管理场景。

符合这里对于配置中心的选型规则。

## 2.2 配置解析选型

配置文件解析选择开源社区内广泛使用的 [viper](https://github.com/spf13/viper)

viper 是目前go项目的最优配置解决方案，它支持以下特性

- setting defaults （支持默认值）
- reading from JSON, TOML, YAML, HCL, envfile and Java properties config files（支持各类文件格式）
- live watching and re-reading of config files (optional) （实时监控配置更新）
- reading from environment variables （读取环境变量）
- reading from remote config systems (etcd or Consul), and watching changes （读取和监听远程配置中心）
- reading from command line flags（读取命令行）
- reading from buffer（读取缓存）
- setting explicit values

Viper 的一个被广为提及的issue 是对字段名称大小写不明感，这在进行项目的迁移的时候容易产生问题，所有在viper 的基础上，这里会引入支持大小写敏感的改造

## 2.3 配置文件类型选型

配置文件类型这里选择go 生态下使用较为广泛的 yaml 文件，对go生态下的各类系统（k8s，traefik网关）等都具有较好的兼容性。

## 3. 流程

```go
func New(config interface{}, opts ...Option) error {
  // 1. 传入一个指针类型的 config ，以实现对配置的动态更新
	if reflect.ValueOf(config).Type().Kind() != reflect.Ptr {
		return errors.New("pointer is required")
	}
  //2. 初始化配置管理器
	o := newOption(opts...)
	manager := &Manager{
		file:   o.file,
		config: config,
		apollo: o.apollo,
	}

  //3. 从本地配置文件读取配置, 并开启一个协程监听文件配置变化
	err := manager.getConfFromFile()
	if err != nil {
		return nil
	}

  //4. 从apollo读取配置, 并开启一个协程监听apollo配置变化
	err = manager.getConfFromApollo()
	return err
}

```

# 4. 使用案例

```go
func main() {
	c := &Config{}
	directory,_ := os.Getwd()
	err := conf.New(c,conf.WithFile(directory+"/example/env.yaml"))
	if err != nil {
		fmt.Println("err",err)
	}
  
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
```



# 5. 完整代码





































