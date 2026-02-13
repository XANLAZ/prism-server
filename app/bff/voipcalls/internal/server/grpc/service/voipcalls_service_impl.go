package service

import (
	"context"

	"github.com/teamgram/proto/mtproto"
	"github.com/teamgram/teamgram-server/app/bff/voipcalls/internal/core"
)

func (s *Service) MessagesDeletePhoneCallHistory(ctx context.Context, request *mtproto.TLMessagesDeletePhoneCallHistory) (*mtproto.Messages_AffectedFoundMessages, error) {
	c := core.New(ctx, s.svcCtx)
	return c.MessagesDeletePhoneCallHistory(request)
}

func (s *Service) PhoneGetCallConfig(ctx context.Context, request *mtproto.TLPhoneGetCallConfig) (*mtproto.DataJSON, error) {
	c := core.New(ctx, s.svcCtx)
	return c.PhoneGetCallConfig(request)
}

func (s *Service) PhoneRequestCall(ctx context.Context, request *mtproto.TLPhoneRequestCall) (*mtproto.Phone_PhoneCall, error) {
	c := core.New(ctx, s.svcCtx)
	return c.PhoneRequestCall(request)
}

func (s *Service) PhoneAcceptCall(ctx context.Context, request *mtproto.TLPhoneAcceptCall) (*mtproto.Phone_PhoneCall, error) {
	c := core.New(ctx, s.svcCtx)
	return c.PhoneAcceptCall(request)
}

func (s *Service) PhoneConfirmCall(ctx context.Context, request *mtproto.TLPhoneConfirmCall) (*mtproto.Phone_PhoneCall, error) {
	c := core.New(ctx, s.svcCtx)
	return c.PhoneConfirmCall(request)
}

func (s *Service) PhoneReceivedCall(ctx context.Context, request *mtproto.TLPhoneReceivedCall) (*mtproto.Bool, error) {
	c := core.New(ctx, s.svcCtx)
	return c.PhoneReceivedCall(request)
}

func (s *Service) PhoneDiscardCall(ctx context.Context, request *mtproto.TLPhoneDiscardCall) (*mtproto.Updates, error) {
	c := core.New(ctx, s.svcCtx)
	return c.PhoneDiscardCall(request)
}

func (s *Service) PhoneSetCallRating(ctx context.Context, request *mtproto.TLPhoneSetCallRating) (*mtproto.Updates, error) {
	c := core.New(ctx, s.svcCtx)
	return c.PhoneSetCallRating(request)
}

func (s *Service) PhoneSaveCallDebug(ctx context.Context, request *mtproto.TLPhoneSaveCallDebug) (*mtproto.Bool, error) {
	c := core.New(ctx, s.svcCtx)
	return c.PhoneSaveCallDebug(request)
}

func (s *Service) PhoneSendSignalingData(ctx context.Context, request *mtproto.TLPhoneSendSignalingData) (*mtproto.Bool, error) {
	c := core.New(ctx, s.svcCtx)
	return c.PhoneSendSignalingData(request)
}

func (s *Service) PhoneSaveCallLog(ctx context.Context, request *mtproto.TLPhoneSaveCallLog) (*mtproto.Bool, error) {
	c := core.New(ctx, s.svcCtx)
	return c.PhoneSaveCallLog(request)
}
