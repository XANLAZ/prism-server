package core

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/teamgram/proto/mtproto"
	"github.com/teamgram/proto/mtproto/rpc/metadata"
	"github.com/teamgram/teamgram-server/app/bff/langpack/internal/svc"
	"github.com/teamgram/teamgram-server/app/bff/bff/client/langpack"
)

type ConfigurationCore struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	MD *metadata.RpcMetadata
}

func New(ctx context.Context, svcCtx *svc.ServiceContext) *ConfigurationCore {
	return &ConfigurationCore{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
		MD:     metadata.RpcMetadataFromIncoming(ctx),
	}
}

func (c *ConfigurationCore) LangpackGetDifference(in *mtproto.TLLangpackGetDifference) (*mtproto.LangPackDifference, error) {
	langCode := in.GetLangCode()
	fromVersion := in.GetFromVersion()
	currentVersion := langpack.GetVersion(langCode)

	c.Logger.Infof("LangpackGetDifference: langCode=%s, fromVersion=%d, currentVersion=%d, isLoaded=%v", 
		langCode, fromVersion, currentVersion, langpack.IsLoaded(langCode))

	// If client has current version, return empty diff
	if fromVersion >= int32(currentVersion) {
		return mtproto.MakeTLLangPackDifference(&mtproto.LangPackDifference{
			LangCode:    langCode,
			FromVersion: fromVersion,
			Version:     int32(currentVersion),
			Strings:     []*mtproto.LangPackString{},
		}).To_LangPackDifference(), nil
	}

	// Return full langpack as difference
	stringsMap := langpack.GetAllStrings(langCode)
	c.Logger.Infof("LangpackGetDifference: stringsMap size=%d", len(stringsMap))

	result := make([]*mtproto.LangPackString, 0, len(stringsMap))
	for key, value := range stringsMap {
		result = append(result, mtproto.MakeTLLangPackString(&mtproto.LangPackString{
			Key:   key,
			Value: value,
		}).To_LangPackString())
	}

	c.Logger.Infof("LangpackGetDifference: result size=%d", len(result))
	return mtproto.MakeTLLangPackDifference(&mtproto.LangPackDifference{
		LangCode:    langCode,
		FromVersion: fromVersion,
		Version:     int32(currentVersion),
		Strings:     result,
	}).To_LangPackDifference(), nil
}

func (c *ConfigurationCore) LangpackGetLangPack(in *mtproto.TLLangpackGetLangPack) (*mtproto.LangPackDifference, error) {
	langCode := in.GetLangCode()
	currentVersion := langpack.GetVersion(langCode)

	c.Logger.Infof("LangpackGetLangPack: langCode=%s, currentVersion=%d, isLoaded=%v", 
		langCode, currentVersion, langpack.IsLoaded(langCode))

	stringsMap := langpack.GetAllStrings(langCode)
	c.Logger.Infof("LangpackGetLangPack: stringsMap size=%d", len(stringsMap))

	result := make([]*mtproto.LangPackString, 0, len(stringsMap))
	for key, value := range stringsMap {
		result = append(result, mtproto.MakeTLLangPackString(&mtproto.LangPackString{
			Key:   key,
			Value: value,
		}).To_LangPackString())
	}

	c.Logger.Infof("LangpackGetLangPack: result size=%d", len(result))
	return mtproto.MakeTLLangPackDifference(&mtproto.LangPackDifference{
		LangCode:    langCode,
		FromVersion: 0,
		Version:     int32(currentVersion),
		Strings:     result,
	}).To_LangPackDifference(), nil
}

func (c *ConfigurationCore) LangpackGetLanguages(in *mtproto.TLLangpackGetLanguages) (*mtproto.Vector_LangPackLanguage, error) {
	return &mtproto.Vector_LangPackLanguage{
		Datas: []*mtproto.LangPackLanguage{
			mtproto.MakeTLLangPackLanguage(&mtproto.LangPackLanguage{
				Name:            "English",
				NativeName:      "English",
				LangCode:        "en",
				PluralCode:      "en",
				StringsCount:    0,
				TranslatedCount: 0,
				TranslationsUrl: "",
			}).To_LangPackLanguage(),
			mtproto.MakeTLLangPackLanguage(&mtproto.LangPackLanguage{
				Name:            "Russian",
				NativeName:      "Русский",
				LangCode:        "ru",
				PluralCode:      "ru",
				StringsCount:    int32(len(langpack.GetAllStrings("ru"))),
				TranslatedCount: int32(len(langpack.GetAllStrings("ru"))),
				TranslationsUrl: "",
			}).To_LangPackLanguage(),
		},
	}, nil
}

func (c *ConfigurationCore) LangpackGetStrings(in *mtproto.TLLangpackGetStrings) (*mtproto.Vector_LangPackString, error) {
	langCode := in.GetLangCode()
	keys := in.GetKeys()

	c.Logger.Infof("LangpackGetStrings: langCode=%s, keys=%v", langCode, keys)

	var result []*mtproto.LangPackString
	stringsMap := langpack.GetAllStrings(langCode)
	for _, key := range keys {
		if val, ok := stringsMap[key]; ok {
			result = append(result, mtproto.MakeTLLangPackString(&mtproto.LangPackString{
				Key:   key,
				Value: val,
			}).To_LangPackString())
		}
	}

	c.Logger.Infof("LangpackGetStrings: result size=%d", len(result))
	return &mtproto.Vector_LangPackString{
		Datas: result,
	}, nil
}

func (c *ConfigurationCore) LangpackGetLanguage(in *mtproto.TLLangpackGetLanguage) (*mtproto.LangPackLanguage, error) {
	_ = in.GetLangCode()

	return mtproto.MakeTLLangPackLanguage(&mtproto.LangPackLanguage{
		Name:            "Russian",
		NativeName:      "Русский",
		LangCode:        "ru",
		PluralCode:      "ru",
		StringsCount:    int32(len(langpack.GetAllStrings("ru"))),
		TranslatedCount: int32(len(langpack.GetAllStrings("ru"))),
		TranslationsUrl: "",
	}).To_LangPackLanguage(), nil
}
