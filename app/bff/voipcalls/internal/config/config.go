package config

import (
	kafka "github.com/teamgram/marmota/pkg/mq"
	"github.com/zeromicro/go-zero/zrpc"
)

type VoipRelayEndpoint struct {
	Id       int64  `json:",optional"`
	Ip       string `json:",optional"`
	Ipv6     string `json:",optional"`
	Port     int32
	PeerTag  string
	Tcp      bool   `json:",optional"`
	Turn     bool   `json:",optional"`
	Stun     bool   `json:",optional"`
	Username string `json:",optional"`
	Password string `json:",optional"`
}

type Config struct {
	zrpc.RpcServerConf
	UserClient zrpc.RpcClientConf
	MsgClient  zrpc.RpcClientConf
	SyncClient *kafka.KafkaProducerConf

	VoipCallConfigJSON string              `json:",optional"`
	VoipRelayEndpoints []VoipRelayEndpoint `json:",optional"`
}
