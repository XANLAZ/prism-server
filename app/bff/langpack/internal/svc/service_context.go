package svc

import (
	"github.com/teamgram/teamgram-server/app/bff/langpack/internal/config"
	"github.com/teamgram/teamgram-server/app/bff/bff/client/langpack"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config  config.Config
	Redis   *redis.Redis
	BizRPC  zrpc.Client
	AuthRPC zrpc.Client
}

func NewServiceContext(c config.Config) *ServiceContext {
	// Load langpacks on service startup
	langpack.LoadAllLangPacks(c.LangpackPath)
	
	return &ServiceContext{
		Config: c,
	}
}
