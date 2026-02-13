package dao

import (
	kafka "github.com/teamgram/marmota/pkg/mq"
	"github.com/teamgram/marmota/pkg/net/rpcx"
	"github.com/teamgram/teamgram-server/app/bff/voipcalls/internal/config"
	msgclient "github.com/teamgram/teamgram-server/app/messenger/msg/msg/client"
	sync_client "github.com/teamgram/teamgram-server/app/messenger/sync/client"
	userclient "github.com/teamgram/teamgram-server/app/service/biz/user/client"
)

type Dao struct {
	userclient.UserClient
	msgclient.MsgClient
	sync_client.SyncClient
}

func New(c config.Config) *Dao {
	return &Dao{
		UserClient: userclient.NewUserClient(rpcx.GetCachedRpcClient(c.UserClient)),
		MsgClient:  msgclient.NewMsgClient(rpcx.GetCachedRpcClient(c.MsgClient)),
		SyncClient: sync_client.NewSyncMqClient(kafka.MustKafkaProducer(c.SyncClient)),
	}
}
