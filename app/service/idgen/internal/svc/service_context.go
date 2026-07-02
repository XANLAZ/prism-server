/*
 * Created from 'scheme.tl' by 'mtprotoc'
 *
 * Copyright (c) 2021-present,  Teamgram Studio (https://teamgram.io).
 *  All rights reserved.
 *
 * Author: teamgramio (teamgram.io@gmail.com)
 */

package svc

import (
	"context"

	"github.com/teamgram/teamgram-server/app/service/idgen/internal/config"
	"github.com/teamgram/teamgram-server/app/service/idgen/internal/dao"
	"github.com/zeromicro/go-zero/core/logx"
)

type ServiceContext struct {
	Config config.Config
	*dao.Dao
}

func NewServiceContext(c config.Config) *ServiceContext {
	svcCtx := &ServiceContext{
		Config: c,
		Dao:    dao.New(c),
	}

	// Sync ID generator counters from MySQL to Redis on startup
	// This prevents duplicate key issues when Redis counters are behind MySQL
	go func() {
		// Give MySQL a moment to be ready if needed
		ctx := context.Background()
		if err := svcCtx.Dao.SyncCountersFromMySQL(ctx); err != nil {
			logx.Errorf("idgen: startup counter sync failed: %v", err)
		} else {
			logx.Info("idgen: startup counter sync completed")
		}
	}()

	return svcCtx
}
