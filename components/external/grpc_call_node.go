package external

import (
	"fmt"
	"math"
	"os"
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

	// "google.golang.org/protobuf/reflect/protoreflect"

	jsoniter "github.com/json-iterator/go"
)

const (
	GRPCReqType  = "_req_type"
	GRPCRespType = "_reply_type"
)

var _ types.Node = (*RPCCallNode)(nil)

// RPCCallNodeConfiguration rpc配置
type RPCCallNodeConfiguration struct {
	PackageName string
	ServiceName string
	Method      string
	Target      string
	// ParamsPattern string
	Headers map[string]string
}

type RPCCallNode struct {
	// 节点配置
	Config RPCCallNodeConfiguration
	// grpc client
	gconn *grpc.ClientConn
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
	x.gconn, err = NewClientConn(x.Config)
	return err
}

func (x *RPCCallNode) OnMsg(ctx types.RuleContext, msg types.RuleMsg) error {
	metaData := msg.Metadata.Values()
	params := str.SprintfDict(msg.Data, metaData)

	req, err := getMessageV1(metaData[GRPCReqType])
	if err != nil {
		ctx.TellFailure(msg, err)
	}
	reply, err := getMessageV1(metaData[GRPCRespType])
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
	ctx.TellSuccess(msg)

	return nil
}

// Destroy 销毁，做一些资源释放操作
func (x *RPCCallNode) Destroy() {
	_ = x.gconn.Close()
}

func NewClientConn(config RPCCallNodeConfiguration) (*grpc.ClientConn, error) {
	dialOption := []grpc.DialOption{}
	conn, err := grpc.Dial(config.Target, DialogOptions(dialOption...)...)
	if err != nil {
		return nil, fmt.Errorf("failed to dial %s %s connect", config.ServiceName, config.Target)
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
	if os.Getenv("RUN_MODE") != "dev" {
		kp := keepalive.ClientParameters{
			Time:                10 * time.Second, // send pings every 10 seconds if there is no activity
			Timeout:             time.Second,      // wait 1 second for ping ack before considering the connection dead
			PermitWithoutStream: true,             // send pings even without active streams
		}
		options = append(options, grpc.WithKeepaliveParams(kp))
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
