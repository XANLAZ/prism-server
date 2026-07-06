package langpack

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/zeromicro/go-zero/core/logx"
)

type LangPack struct {
	Version   int               `json:"version"`
	LangCode  string            `json:"lang_code"`
	Strings   map[string]string `json:"strings"`
}

var (
	langPacks = make(map[string]*LangPack)
	mu        sync.RWMutex
	loaded    = false
)

func LoadLangPack(langCode, basePath string) error {
	mu.Lock()
	defer mu.Unlock()

	if langPacks[langCode] != nil {
		return nil // already loaded
	}

	filePath := filepath.Join(basePath, langCode+".json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		logx.Errorf("Failed to read langpack %s: %v", filePath, err)
		return err
	}

	var lp LangPack
	if err := json.Unmarshal(data, &lp); err != nil {
		logx.Errorf("Failed to parse langpack %s: %v", filePath, err)
		return err
	}

	langPacks[langCode] = &lp
	logx.Infof("Loaded langpack: %s (version %d, %d strings)", langCode, lp.Version, len(lp.Strings))
	return nil
}

func LoadAllLangPacks(basePath string) error {
	mu.Lock()
	defer mu.Unlock()

	if loaded {
		return nil
	}

	langpacksDir := basePath
	entries, err := os.ReadDir(langpacksDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if filepath.Ext(entry.Name()) == ".json" {
			langCode := entry.Name()[:len(entry.Name())-5] // remove .json
			filePath := filepath.Join(langpacksDir, entry.Name())
			data, err := os.ReadFile(filePath)
			if err != nil {
				logx.Errorf("Failed to read langpack %s: %v", filePath, err)
				continue
			}

			var lp LangPack
			if err := json.Unmarshal(data, &lp); err != nil {
				logx.Errorf("Failed to parse langpack %s: %v", filePath, err)
				continue
			}

			langPacks[langCode] = &lp
			logx.Infof("Loaded langpack: %s (version %d, %d strings)", langCode, lp.Version, len(lp.Strings))
		}
	}

	loaded = true
	return nil
}

func GetString(langCode, key string) string {
	mu.RLock()
	defer mu.RUnlock()

	if lp, ok := langPacks[langCode]; ok {
		if val, ok := lp.Strings[key]; ok {
			return val
		}
	}
	// fallback to English (empty string)
	return ""
}

func GetAllStrings(langCode string) map[string]string {
	mu.RLock()
	defer mu.RUnlock()

	if lp, ok := langPacks[langCode]; ok {
		return lp.Strings
	}
	return map[string]string{}
}

func GetVersion(langCode string) int {
	mu.RLock()
	defer mu.RUnlock()

	if lp, ok := langPacks[langCode]; ok {
		return lp.Version
	}
	return 0
}

func IsLoaded(langCode string) bool {
	mu.RLock()
	defer mu.RUnlock()
	_, ok := langPacks[langCode]
	return ok
}
