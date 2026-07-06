/*
 * WARNING! All changes made in this file will be lost!
 * Created from 'scheme.tl' by 'mtprotoc'
 *
 * Copyright 2024 Teamgram Authors.
 *  All rights reserved.
 *
 * Author: teamgramio (teamgram.io@gmail.com)
 */

package service

import (
	"context"

	"github.com/teamgram/proto/mtproto"
	"github.com/teamgram/teamgram-server/app/bff/langpack/internal/core"
)

// LangpackGetDifference
// langpack.getDifference#42c6978f lang_code:string from_version:int = LangPackDifference;
func (s *Service) LangpackGetDifference(ctx context.Context, request *mtproto.TLLangpackGetDifference) (*mtproto.LangPackDifference, error) {
	c := core.New(ctx, s.svcCtx)
	c.Logger.Debugf("langpack.getDifference - metadata: {%s}, request: {%s}", c.MD, request)

	r, err := c.LangpackGetDifference(request)
	if err != nil {
		return nil, err
	}

	c.Logger.Debugf("langpack.getDifference - reply: {%s}", r)
	return r, err
}

// LangpackGetLangPack
// langpack.getLangPack#3e6b547a lang_code:string = LangPackDifference;
func (s *Service) LangpackGetLangPack(ctx context.Context, request *mtproto.TLLangpackGetLangPack) (*mtproto.LangPackDifference, error) {
	c := core.New(ctx, s.svcCtx)
	c.Logger.Debugf("langpack.getLangPack - metadata: {%s}, request: {%s}", c.MD, request)

	r, err := c.LangpackGetLangPack(request)
	if err != nil {
		return nil, err
	}

	c.Logger.Debugf("langpack.getLangPack - reply: {%s}", r)
	return r, err
}

// LangpackGetLanguages
// langpack.getLanguages#42c6978f = Vector<LangPackLanguage>;
func (s *Service) LangpackGetLanguages(ctx context.Context, request *mtproto.TLLangpackGetLanguages) (*mtproto.Vector_LangPackLanguage, error) {
	c := core.New(ctx, s.svcCtx)
	c.Logger.Debugf("langpack.getLanguages - metadata: {%s}, request: {%s}", c.MD, request)

	r, err := c.LangpackGetLanguages(request)
	if err != nil {
		return nil, err
	}

	c.Logger.Debugf("langpack.getLanguages - reply: {%s}", r)
	return r, err
}

// LangpackGetStrings
// langpack.getStrings#7a0aec79 lang_code:string keys:Vector<string> = Vector<LangPackString>;
func (s *Service) LangpackGetStrings(ctx context.Context, request *mtproto.TLLangpackGetStrings) (*mtproto.Vector_LangPackString, error) {
	c := core.New(ctx, s.svcCtx)
	c.Logger.Debugf("langpack.getStrings - metadata: {%s}, request: {%s}", c.MD, request)

	r, err := c.LangpackGetStrings(request)
	if err != nil {
		return nil, err
	}

	c.Logger.Debugf("langpack.getStrings - reply: {%s}", r)
	return r, err
}

// LangpackGetLanguage
// langpack.getLanguage#42c6978f lang_code:string = LangPackLanguage;
func (s *Service) LangpackGetLanguage(ctx context.Context, request *mtproto.TLLangpackGetLanguage) (*mtproto.LangPackLanguage, error) {
	c := core.New(ctx, s.svcCtx)
	c.Logger.Debugf("langpack.getLanguage - metadata: {%s}, request: {%s}", c.MD, request)

	r, err := c.LangpackGetLanguage(request)
	if err != nil {
		return nil, err
	}

	c.Logger.Debugf("langpack.getLanguage - reply: {%s}", r)
	return r, err
}
