/*
 * WARNING! All changes made in this file will be lost!
 * Created from 'scheme.tl' by 'mtprotoc'
 *
 * Copyright 2022 Teamgram Authors.
 *  All rights reserved.
 *
 * Author: teamgramio (teamgram.io@gmail.com)
 */

package langpack_helper

import (
	"github.com/teamgram/teamgram-server/app/bff/langpack/internal/config"
	"github.com/teamgram/teamgram-server/app/bff/langpack/internal/server/grpc/service"
	"github.com/teamgram/teamgram-server/app/bff/langpack/internal/svc"
	"github.com/teamgram/teamgram-server/app/bff/bff/client/langpack"
)

type (
	Config = config.Config
)

func New(c Config) *service.Service {
	return service.New(svc.NewServiceContext(c))
}

var (
	LoadAllLangPacks = langpack.LoadAllLangPacks
	LoadLangPack     = langpack.LoadLangPack
	GetString        = langpack.GetString
	GetAllStrings    = langpack.GetAllStrings
	GetVersion       = langpack.GetVersion
	IsLoaded         = langpack.IsLoaded
)
