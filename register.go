package gecko

import (
	"container/list"
	"github.com/parkingwang/go-conf"
	"github.com/yoojia/go-gecko/utils"
)

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

// 负责对Engine组件的注册管理
type Registration struct {
	// 组件管理
	outputsMap   map[string]OutputDevice
	inputsMap    map[string]InputDevice
	decodersMap  map[string]Decoder
	encodersMap  map[string]Encoder
	plugins      *list.List
	interceptors *list.List
	drivers      *list.List
	outputs      *list.List
	inputs       *list.List
	// Hooks
	startBeforeHooks *list.List
	startAfterHooks  *list.List
	stopBeforeHooks  *list.List
	stopAfterHooks   *list.List
	// 组件创建工厂函数
	factories map[string]BundleFactory
}

func prepare() *Registration {
	re := new(Registration)
	re.outputsMap = make(map[string]OutputDevice)
	re.inputsMap = make(map[string]InputDevice)
	re.decodersMap = make(map[string]Decoder)
	re.encodersMap = make(map[string]Encoder)
	re.plugins = list.New()
	re.interceptors = list.New()
	re.drivers = list.New()
	re.inputs = list.New()
	re.outputs = list.New()
	re.startBeforeHooks = list.New()
	re.startAfterHooks = list.New()
	re.stopBeforeHooks = list.New()
	re.stopAfterHooks = list.New()
	re.factories = make(map[string]BundleFactory)
	return re
}

// 添加Encoder
func (re *Registration) AddEncoder(name string, encoder Encoder) {
	if _, ok := re.encodersMap[name]; ok {
		ZapSugarLogger.Panicw("Encoder类型重复", "type", name)
	} else {
		re.encodersMap[name] = encoder
	}
}

// 添加Decoder
func (re *Registration) AddDecoder(name string, decoder Decoder) {
	if _, ok := re.decodersMap[name]; ok {
		ZapSugarLogger.Panicw("Decoder类型重复", "type", name)
	} else {
		re.decodersMap[name] = decoder
	}
}

// 添加OutputDevice
func (re *Registration) AddOutputDevice(device OutputDevice) {
	uuid := re.ensureUniqueUUID(device.GetAddress().UUID)
	re.outputsMap[uuid] = device
	re.outputs.PushBack(device)
}

// 添加InputDevice
func (re *Registration) AddInputDevice(device InputDevice) {
	uuid := re.ensureUniqueUUID(device.GetAddress().UUID)
	re.inputsMap[uuid] = device
	re.inputs.PushBack(device)
}

// 添加Plugin
func (re *Registration) AddPlugin(plugin Plugin) {
	re.plugins.PushBack(plugin)
}

// 添加Interceptor
func (re *Registration) AddInterceptor(interceptor Interceptor) {
	re.interceptors.PushBack(interceptor)
}

// 添加Driver
func (re *Registration) AddDriver(driver Driver) {
	re.drivers.PushBack(driver)
}

func (re *Registration) AddStartBeforeHook(hook HookFunc) {
	re.startBeforeHooks.PushBack(hook)
}

func (re *Registration) AddStartAfterHook(hook HookFunc) {
	re.startAfterHooks.PushBack(hook)
}

func (re *Registration) AddStopBeforeHook(hook HookFunc) {
	re.stopBeforeHooks.PushBack(hook)
}

func (re *Registration) AddStopAfterHook(hook HookFunc) {
	re.startAfterHooks.PushBack(hook)
}

func (re *Registration) showBundles() {
	zlog := ZapSugarLogger
	zlog.Infof("已加载 Interceptors: %d", re.interceptors.Len())
	utils.ForEach(re.interceptors, func(it interface{}) {
		zlog.Info("  - Interceptor: " + utils.GetClassName(it))
	})

	zlog.Infof("已加载 InputDevices: %d", re.inputs.Len())
	utils.ForEach(re.inputs, func(it interface{}) {
		zlog.Info("  - InputDevice: " + utils.GetClassName(it))
	})

	zlog.Infof("已加载OutputDevices: %d", re.outputs.Len())
	utils.ForEach(re.outputs, func(it interface{}) {
		zlog.Info("  - OutputDevice: " + utils.GetClassName(it))
	})

	zlog.Infof("已加载 Drivers: %d", re.drivers.Len())
	utils.ForEach(re.drivers, func(it interface{}) {
		zlog.Info("  - Driver: " + utils.GetClassName(it))
	})

	zlog.Infof("已加载 Plugins: %d", re.plugins.Len())
	utils.ForEach(re.plugins, func(it interface{}) {
		zlog.Info("  - Plugin: " + utils.GetClassName(it))
	})
}

// Deprecated: Use AddBundleFactory instead.
func (re *Registration) RegisterBundleFactory(typeName string, factory BundleFactory) {
	re.AddBundleFactory(typeName, factory)
}

// Deprecated: Use AddCodecFactory instead.
func (re *Registration) RegisterCodecFactory(typeName string, factory CodecFactory) {
	re.AddCodecFactory(typeName, factory)
}

// 注册组件工厂函数
func (re *Registration) AddBundleFactory(typeName string, factory BundleFactory) {
	zlog := ZapSugarLogger
	if _, ok := re.factories[typeName]; ok {
		zlog.Warnf("组件类型[%s]，旧的工厂函数将被覆盖为： %s", typeName, utils.GetClassName(factory))
	}
	zlog.Infof("正在注册组件工厂函数： %s", typeName)
	re.factories[typeName] = factory
}

// 注册编码解码工厂函数
func (re *Registration) AddCodecFactory(typeName string, factory CodecFactory) {
	codec := factory()
	switch codec.(type) {
	case Decoder:
		re.AddDecoder(typeName, codec.(Decoder))

	case Encoder:
		re.AddEncoder(typeName, codec.(Encoder))

	default:
		ZapSugarLogger.Panicf("未知的编/解码类型[%s]，工厂函数： %s", typeName, utils.GetClassName(factory))
	}
}

// 查找指定类型的
func (re *Registration) findFactory(typeName string) (BundleFactory, bool) {
	if f, ok := re.factories[typeName]; ok {
		return f, true
	} else {
		return nil, false
	}
}

func (re *Registration) ensureUniqueUUID(uuid string) string {
	zlog := ZapSugarLogger
	if _, ok := re.inputsMap[uuid]; ok {
		zlog.Panicf("设备UUID重复[Input]：%s", uuid)
	} else if _, ok := re.outputsMap[uuid]; ok {
		zlog.Panicf("设备UUID重复[Output]：%s", uuid)
	}
	return uuid
}

// 注册组件，如果注册失败，返回False
func (re *Registration) registerIfHit(configs *cfg.Config, initFunc func(bundle Initialize, args *cfg.Config)) bool {
	if configs.IsEmpty() {
		return false
	}
	zlog := ZapSugarLogger
	configs.ForEach(func(bundleType string, item interface{}) {
		asMap, ok := item.(map[string]interface{})
		if !ok {
			zlog.Panicf("组件配置信息类型错误: %s", bundleType)
		}
		config := cfg.Wrap(asMap)
		if config.MustBool("disable") {
			zlog.Infof("组件[%s]在配置中禁用", bundleType)
			return
		}

		// 配置选项中，指定 type 字段为类型名称
		if typeName := config.MustString("type"); "" != typeName {
			bundleType = typeName
		}

		factory, ok := re.findFactory(bundleType)
		if !ok {
			zlog.Panicf("组件类型[%s]，没有注册对应的工厂函数", bundleType)
		}
		// 根据类型注册
		bundle := factory()
		switch bundle.(type) {

		case Driver:
			re.AddDriver(bundle.(Driver))

		case Interceptor:
			it := bundle.(Interceptor)
			it.setPriority(int(config.MustInt64("priority")))
			re.AddInterceptor(it)

		case VirtualDevice:
			device := bundle.(VirtualDevice)
			if name := config.MustString("name"); "" == name {
				zlog.Panicf("VirtualDevice[%s]配置项[name]是必填参数", bundleType)
			} else {
				device.setName(name)
			}

			address := DeviceAddress{
				UUID:    config.MustString("uuid"),
				Group:   config.MustString("group"),
				Private: config.MustString("private"),
			}
			if !address.IsValid() {
				zlog.Panicf("VirtualDevice[%s]配置项[uuid/group/private]是必填参数", bundleType)
			}
			device.setAddress(address)

			if name := config.MustString("encoder"); "" == name {
				if nil == device.GetEncoder() {
					zlog.Panicf("未设置默认Encoder时，Device[%s]配置项[encoder]是必填参数", bundleType)
				}
			} else {
				if encoder, ok := re.encodersMap[name]; ok {
					device.setEncoder(encoder)
				} else {
					zlog.Panicf("Encoder[%s]未注册", name)
				}
			}

			if name := config.MustString("decoder"); "" == name {
				if nil == device.GetDecoder() {
					zlog.Panicf("未设置默认Decoder时，Device[%s]配置项[decoder]是必填参数", bundleType)
				}
			} else {
				if decoder, ok := re.decodersMap[name]; ok {
					device.setDecoder(decoder)
				} else {
					zlog.Panicf("Decoder[%s]未注册", name)
				}
			}

			if inputDevice, ok := device.(InputDevice); ok {
				if topic := config.MustString("topic"); "" == topic {
					zlog.Panicf("Device[%s]配置项[topic]是必填参数", bundleType)
				} else {
					inputDevice.setTopic(topic)
				}
				re.AddInputDevice(inputDevice)
			} else if outputDevice, ok := device.(OutputDevice); ok {
				re.AddOutputDevice(outputDevice)
			} else {
				zlog.Panicf("未知VirtualDevice类型： %s", utils.GetClassName(device))
			}

		default:
			if plg, ok := bundle.(Plugin); ok {
				re.AddPlugin(plg)
			} else {
				zlog.Panicf("未支持的组件类型：%s. 你是否没有实现某个函数接口？", bundleType)
			}
		}

		// 需要Topic过滤
		if tf, ok := bundle.(NeedTopicFilter); ok {
			if topics, err := config.MustStringArray("topics"); nil != err || 0 == len(topics) {
				zlog.Panicw("配置项中[topics]必须是字符串数组", "type", bundleType, "error", err)
			} else {
				tf.setTopics(topics)
			}
		}

		// 组件初始化。由外部函数处理，减少不必要的依赖
		if init, ok := bundle.(Initialize); ok {
			initFunc(init, config.MustConfig("InitArgs"))
		}
	})
	return true
}
