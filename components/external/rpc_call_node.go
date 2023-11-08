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

	jsoniter "github.com/json-iterator/go"
)

var _ types.Node = (*RPCCallNode)(nil)

// RPCCallNodeConfiguration rpc配置
type RPCCallNodeConfiguration struct {
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
	return "rpcCall"
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
	paramsMap := make(map[string]interface{})
	err := jsoniter.UnmarshalFromString(params, &paramsMap)
	if err != nil {
		ctx.TellFailure(msg, err)
	}
	retMap := make(map[string]interface{})

	// 设置header
	gctx := ctx.GetContext()
	for key, value := range x.Config.Headers {
		header := make(map[string]string)
		header[str.SprintfDict(key, metaData)] = str.SprintfDict(value, metaData)
		md := metadata.New(header)
		gctx = metadata.NewIncomingContext(gctx, md)
	}
	err = x.gconn.Invoke(gctx, x.Config.Method, &paramsMap, &retMap)
	if err != nil {
		ctx.TellFailure(msg, err)
	}
	// succeed
	data, err := jsoniter.MarshalToString(retMap)
	if err != nil {
		ctx.TellFailure(msg, err)
	}
	msg.Data = data

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

// func UnaryClientInterceptor(mdw ...grpc.UnaryClientInterceptor) grpc.DialOption {
// 	chain := []grpc.UnaryClientInterceptor{
// 		trace.UnaryClientInterceptor(),
// 		requestid.UnaryClientInterceptor(),
// 		logmdw.UnaryClientInterceptor(log.FromContext),
// 	}
// 	chain = append(chain, mdw...)
// 	return grpc.WithChainUnaryInterceptor(chain...)
// }

// func DiscoveryResolver(discovery registry.Discovery, scheme string) grpc.DialOption {
// 	return grpc.WithResolvers(lgrpc.NewResolverBuilder(discovery, scheme))
// }

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
