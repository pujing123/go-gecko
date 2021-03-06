package gecko

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

// Driver-用户驱动，是实现设备与设备之间联动、设备事件响应业务处理的核心组件，它通常与Interceptor一起完成某种业务功能；
// 它负责监听接收特定事件Topic的设备事件，经过内部数据库、业务方法等逻辑计算后，控制OutputDeliverer来操作下一级输出设备。
// 最典型的例子是：Driver接收到门禁刷卡ID后，驱动门锁开关设备；
type Driver interface {
	NeedTopicFilter
	NeedName
	// 处理外部请求，返回响应结果。
	// 在Driver内部，可以通过 OutputDeliverer 来控制其它设备。
	Drive(attrs Attributes, topic string, uuid string, in *MessagePacket, fn OutputDeliverer, ctx Context) (out *MessagePacket, err error)
}

//// Driver抽象实现

type AbcDriver struct {
	Driver
	name   string
	topics []*TopicExpr
}

func (ad *AbcDriver) setName(name string) {
	ad.name = name
}

// 获取Driver名字
func (ad *AbcDriver) GetName() string {
	return ad.name
}

func (ad *AbcDriver) setTopics(topics []string) {
	for _, t := range topics {
		ad.topics = append(ad.topics, newTopicExpr(t))
	}
}

// 获取Driver可处理的Topic列表
func (ad *AbcDriver) GetTopicExpr() []*TopicExpr {
	return ad.topics
}

func NewAbcDriver() *AbcDriver {
	return &AbcDriver{
		topics: make([]*TopicExpr, 0),
	}
}
