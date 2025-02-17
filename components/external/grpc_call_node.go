package external

import (
	"fmt"
	"io"
	"math"
	"sync"
	"time"

	"github.com/xyzbit/rulego/api/types"
	"github.com/xyzbit/rulego/utils/maps"
	"github.com/xyzbit/rulego/utils/str"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/runtime/protoiface"

	jsoniter "github.com/json-iterator/go"
)

var connCache sync.Map

func init() {
	Registry.Add(&RPCCallNode{})
}

var _ types.Node = (*RPCCallNode)(nil)

// RPCCallNodeConfiguration rpc配置
type RPCCallNodeConfiguration struct {
	Method   string
	ReqType  string
	RespType string
	Target   string
	// ParamsPattern string
	Headers   map[string]string
	KeepAlive bool
}

type ICloseableClientConn interface {
	grpc.ClientConnInterface
	io.Closer
}

type RPCCallNode struct {
	// 节点配置
	Config RPCCallNodeConfiguration
	// grpc client
	gconn ICloseableClientConn
}

// 实现Node接口
func (x *RPCCallNode) New() types.Node {
	return &RPCCallNode{}
}

func (x *RPCCallNode) Type() string {
	return "grpcCall"
}

func (x *RPCCallNode) Init(ruleConfig types.Config, configuration types.Configuration) error {
	err := maps.Map2Struct(configuration, &x.Config)
	if err != nil {
		return err
	}
	gconn, ok := connCache.Load(x.Config.Target)
	if ok {
		x.gconn = gconn.(*grpc.ClientConn)
		return nil
	}

	x.gconn, err = NewClientConn(x.Config)
	if err != nil {
		return err
	}
	connCache.Store(x.Config.Target, x.gconn)
	return nil
}

func (x *RPCCallNode) OnMsg(ctx types.RuleContext, msg types.RuleMsg) error {
	metaData := msg.Metadata.Values()
	params := str.SprintfDict(msg.Data, metaData)

	req, err := getMessageV1(x.Config.ReqType)
	if err != nil {
		ctx.TellFailure(msg, err)
	}
	reply, err := getMessageV1(x.Config.RespType)
	if err != nil {
		ctx.TellFailure(msg, err)
	}
	if err := jsoniter.Unmarshal([]byte(params), req); err != nil {
		ctx.TellFailure(msg, err)
	}

	gctx := ctx.GetContext()
	for key, value := range x.Config.Headers {
		header := make(map[string]string)
		header[str.SprintfDict(key, metaData)] = str.SprintfDict(value, metaData)
		md := metadata.New(header)
		gctx = metadata.NewIncomingContext(gctx, md)
	}

	err = x.gconn.Invoke(gctx, x.Config.Method, req, reply)
	if err != nil {
		ctx.TellFailure(msg, err)
	}
	data, err := jsoniter.MarshalToString(reply)
	if err != nil {
		ctx.TellFailure(msg, err)
	}
	msg.Data = data
	msg.Metadata.PutValue(x.Config.RespType, data)
	ctx.TellSuccess(msg)
	return nil
}

// Destroy 销毁，做一些资源释放操作
func (x *RPCCallNode) Destroy() {
	_ = x.gconn.Close()
}

func NewClientConn(config RPCCallNodeConfiguration) (*grpc.ClientConn, error) {
	dialOption := []grpc.DialOption{}
	if config.KeepAlive {
		keepaliveParams := keepalive.ClientParameters{
			Time:                10 * time.Second,
			Timeout:             3 * time.Second,
			PermitWithoutStream: true,
		}
		dialOption = append(dialOption, grpc.WithKeepaliveParams(keepaliveParams))
	}

	conn, err := grpc.Dial(config.Target, DialogOptions(dialOption...)...)
	if err != nil {
		return nil, fmt.Errorf("failed to dial %s connect", config.Target)
	}
	return conn, nil
}

func DialogOptions(opts ...grpc.DialOption) []grpc.DialOption {
	options := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithWriteBufferSize(1024 * 1024),
		grpc.WithReadBufferSize(4096 * 1024),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(math.MaxInt32),
			grpc.MaxCallSendMsgSize(math.MaxInt32),
		),
		grpc.WithDefaultServiceConfig(`{"loadBalancingConfig": [{"round_robin":{}}]}`),
	}

	options = append(options, opts...)
	return options
}

func getMessageV1(messageType string) (protoiface.MessageV1, error) {
	if messageType == "" {
		return nil, fmt.Errorf("message type is empty")
	}
	messageDescriptor, err := protoregistry.GlobalTypes.FindMessageByName(protoreflect.FullName(messageType))
	if err != nil {
		return nil, fmt.Errorf("unknown message type: %s", messageType)
	}

	protoMessage := messageDescriptor.New().Interface()
	messageV1, ok := protoMessage.(protoiface.MessageV1)
	if !ok {
		return nil, fmt.Errorf("message type is not protoiface.MessageV1: %s", messageType)
	}

	return messageV1, nil
}
