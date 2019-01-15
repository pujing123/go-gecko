package udp

import (
	"context"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/yoojia/go-gecko"
	"net"
	"parkingwang.com/go-conf"
	"time"
)

func NewAbcUdpReaderDevice() *AbcUdpReaderDevice {
	return &AbcUdpReaderDevice{
		AbcInputDevice: gecko.NewAbcInputDevice(),
	}
}

// UDP服务器读取设备
type AbcUdpReaderDevice struct {
	*gecko.AbcInputDevice
	maxBufferSize  int64
	readTimeout    time.Duration
	cancelCtx      context.Context
	cancelFun      context.CancelFunc
	onServeHandler func(bytes []byte, ctx gecko.Context, deliverer gecko.Deliverer) error
}

func (ur *AbcUdpReaderDevice) OnInit(args map[string]interface{}, ctx gecko.Context) {
	config := conf.WrapImmutableMap(args)
	ur.maxBufferSize = config.GetInt64OrDefault("bufferSizeKB", 1) * 1024
	ur.readTimeout = config.GetDurationOrDefault("readTimeout", time.Second*10)
}

func (ur *AbcUdpReaderDevice) OnStart(ctx gecko.Context) {
	address := ur.GetUnionAddress()
	ur.withTag(log.Info).Msgf("使用UDPReader，服务地址： %s", address)
	ur.cancelCtx, ur.cancelFun = context.WithCancel(context.Background())
}

func (ur *AbcUdpReaderDevice) OnStop(ctx gecko.Context) {
	ur.cancelFun()
}

func (ur *AbcUdpReaderDevice) Serve(ctx gecko.Context, deliverer gecko.Deliverer) error {
	if nil == ur.onServeHandler {
		return errors.New("未设置onServeHandler接口")
	}
	address := ur.GetUnionAddress()
	ur.withTag(log.Info).Msgf("绑定UDP服务端： %s", address)
	conn, cErr := net.ListenPacket("udp", address)
	if nil != cErr {
		return cErr
	}
	defer conn.Close()

	buffer := make([]byte, ur.maxBufferSize)
	for {
		select {
		case <-ur.cancelCtx.Done():
			return nil

		default:
			if err := conn.SetReadDeadline(time.Now().Add(ur.readTimeout)); nil != err {
				return err
			}
			n, _, err := conn.ReadFrom(buffer)
			if nil != err {
				return err
			}
			frame := gecko.NewPackFrame(buffer[:n])
			if err := ur.onServeHandler(frame, ctx, deliverer); nil != err {
				return err
			}
		}
	}
	return nil
}

// 设置Serve处理函数
func (ur *AbcUdpReaderDevice) SetServeHandler(handler func([]byte, gecko.Context, gecko.Deliverer) error) {
	ur.onServeHandler = handler
}

func (ur *AbcUdpReaderDevice) withTag(f func() *zerolog.Event) *zerolog.Event {
	return f().Str("tag", "AbcUdpReaderDevice")
}
