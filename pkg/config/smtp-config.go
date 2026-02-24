package config

import (
	"github.com/spf13/viper"
)

const (
	SmtpConfigKey = "smtp"
)

type SmtpConfig struct {
	Upstream UpstreamConfig
}

type UpstreamConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

func (s *SmtpConfig) Load(v *viper.Viper) error {
	if !v.IsSet(SmtpConfigKey) {
		return ErrModConfigNotFound
	}
	if err := v.UnmarshalKey(SmtpConfigKey, s); err != nil {
		return err
	}
	return nil
}

func (s *SmtpConfig) Name() string {
	return SmtpConfigKey
}

func GetSmtpConfig() (*SmtpConfig, error) {
	m := GetModule(SmtpConfigKey)
	if m == nil {
		return nil, ErrLoadModule
	}
	if val, ok := m.(*SmtpConfig); ok {
		return val, nil
	}
	return nil, ErrModuleAssertion
}
