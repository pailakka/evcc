package settings

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/evcc-io/evcc/core/keys"
	dbsettings "github.com/evcc-io/evcc/server/db/settings"
	"github.com/evcc-io/evcc/util"
	"github.com/evcc-io/evcc/util/config"
	"github.com/spf13/cast"
)

var _ Settings = (*ConfigSettings)(nil)

type ConfigSettings struct {
	mu   sync.Mutex
	log  *util.Logger
	conf *config.Config
}

func NewConfigSettingsAdapter(log *util.Logger, conf *config.Config) *ConfigSettings {
	return &ConfigSettings{log: log, conf: conf}
}

func (s *ConfigSettings) get(key string) (any, error) {
	if s.externalCompatibilityKey(key) != "" {
		return dbsettings.String(s.externalCompatibilityKey(key))
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	val := s.conf.Named().Other[key]
	if val == nil {
		return nil, errors.New("not found")
	}
	return val, nil
}

// TODO remove broken error handling when settings api is retired
func (s *ConfigSettings) set(key string, val any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	data := s.conf.Named().Other
	if externalKey := s.externalCompatibilityKey(key); externalKey != "" {
		switch v := val.(type) {
		case string:
			dbsettings.SetString(externalKey, v)
		case float64:
			dbsettings.SetFloat(externalKey, v)
		default:
			dbsettings.SetString(externalKey, cast.ToString(v))
		}

		delete(data, key)
	} else {
		data[key] = val
	}
	if err := s.conf.Update(data); err != nil {
		s.log.ERROR.Println(err)
	}
}

func (s *ConfigSettings) externalCompatibilityKey(key string) string {
	switch key {
	case keys.PlanContribution, keys.PlanSupportMode:
		return fmt.Sprintf("cfg.%s.%d.%s", s.conf.Class, s.conf.ID, key)
	default:
		return ""
	}
}

func (s *ConfigSettings) SetString(key string, val string) {
	s.set(key, val)
}

func (s *ConfigSettings) SetInt(key string, val int64) {
	s.set(key, val)
}

func (s *ConfigSettings) SetFloat(key string, val float64) {
	s.set(key, val)
}

func (s *ConfigSettings) SetFloatPtr(key string, val *float64) {
	s.set(key, val)
}

func (s *ConfigSettings) SetTime(key string, val time.Time) {
	s.set(key, val)
}

func (s *ConfigSettings) SetBool(key string, val bool) {
	s.set(key, val)
}

func (s *ConfigSettings) SetJson(key string, val any) error {
	s.set(key, val)
	return nil
}

func (s *ConfigSettings) String(key string) (string, error) {
	val, err := s.get(key)
	if err != nil {
		return "", err
	}
	return cast.ToStringE(val)
}

func (s *ConfigSettings) Int(key string) (int64, error) {
	val, err := s.get(key)
	if err != nil {
		return 0, err
	}
	return cast.ToInt64E(val)
}

func (s *ConfigSettings) Float(key string) (float64, error) {
	val, err := s.get(key)
	if err != nil {
		return 0, err
	}
	return cast.ToFloat64E(val)
}

func (s *ConfigSettings) Time(key string) (time.Time, error) {
	val, err := s.get(key)
	if err != nil {
		return time.Time{}, err
	}
	return cast.ToTimeE(val)
}

func (s *ConfigSettings) Bool(key string) (bool, error) {
	val, err := s.get(key)
	if err != nil {
		return false, err
	}
	return cast.ToBoolE(val)
}

func (s *ConfigSettings) Json(key string, res any) error {
	val, err := s.get(key)
	if err != nil {
		return err
	}

	switch v := val.(type) {
	case string:
		if v == "" {
			return err
		}
		return json.Unmarshal([]byte(v), &res)
	default:
		data, err := json.Marshal(v)
		if err != nil {
			return err
		}
		return json.Unmarshal(data, &res)
	}
}
