package voipcalls_helper

import (
	"github.com/teamgram/teamgram-server/app/bff/voipcalls/internal/config"
	"github.com/teamgram/teamgram-server/app/bff/voipcalls/internal/server/grpc/service"
	"github.com/teamgram/teamgram-server/app/bff/voipcalls/internal/svc"
)

type (
	Config            = config.Config
	VoipRelayEndpoint = config.VoipRelayEndpoint
)

func New(c Config) *service.Service {
	return service.New(svc.NewServiceContext(c))
}
