package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/spf13/viper"
	"github.com/xyroscar/common-lib/pkg/logger"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

var (
	ErrConfigNotFound       error = errors.New("config not found")
	ErrUnmarshalConfig      error = errors.New("failed to unmarshal config")
	ErrLoadModule           error = errors.New("failed to load module")
	ErrAppNameNotConfigured error = errors.New("app name is not configured")

	ErrModConfigNotFound error = errors.New("module config not found")
	ErrModuleAssertion   error = errors.New("error asserting module type")
)

var (
	config     *Config
	configOnce sync.Once
	v          *viper.Viper
	appName    string
)

func SetAppName(a string) {
	appName = a
}

var (
	modMu   sync.RWMutex
	modules = make(map[string]Module)
)

func RegisterModule(module Module) {
	modMu.Lock()
	defer modMu.Unlock()
	modules[module.Name()] = module
}

type Config struct {
	AppName       string
	BaseDir       string
	LoggingConfig LoggingConfig
	HostConfig    HostConfig
}

type HostConfig struct {
	Host string
	Port int
}

type LoggingConfig struct {
	Level       string
	MaxFileSize int
	MaxAge      int
	MaxNumFiles int
	Compress    bool
}

type Module interface {
	Load(v *viper.Viper) error

	Name() string
}

func GetConfig() *Config {
	configOnce.Do(func() {
		err := InitAppConfig()
		if err != nil {
			log.Println("Failed to initialize app config")
			if errors.Is(err, ErrConfigNotFound) {
				DefaultConfig()
				err = InitAppConfig()
				if err != nil {
					log.Println("Failed to initialize app config")
					os.Exit(1)
				}
			} else {
				os.Exit(1)
			}
		}
	})
	return config
}

func InitAppConfig() error {
	if appName == "" {
		logger.Error("App Name cannot be empty")
		return ErrConfigNotFound
	}
	v = viper.New()

	v.AutomaticEnv()
	a := strings.ToUpper(appName)
	v.SetEnvPrefix(a)

	configPathVar := fmt.Sprintf("%s_CONFIG_PATH", a)
	if os.Getenv(configPathVar) != "" {
		v.SetConfigFile(os.Getenv(configPathVar))
	} else {
		v.SetConfigName("config")
		v.AddConfigPath(fmt.Sprintf("/opt/%s/config", strings.ToLower(appName)))
	}
	v.SetConfigType("yaml")
	if err := v.ReadInConfig(); err != nil {
		logger.Error("Error reading configs", zap.Error(err))
		return ErrConfigNotFound
	}

	if err := v.Unmarshal(&config); err != nil {
		return ErrUnmarshalConfig
	}

	modMu.RLock()
	defer modMu.RUnlock()
	for _, module := range modules {
		logger.Debug("Loading module", zap.String("module", module.Name()))
		if err := module.Load(v); err != nil {
			logger.Error("Error loading module", zap.Error(err))
			return ErrLoadModule
		}
	}
	return nil
}

func DefaultConfig() *Config {
	baseDir := fmt.Sprintf("/opt/%s", strings.ToLower(appName))
	c := &Config{
		AppName: appName,
		BaseDir: baseDir,
		LoggingConfig: LoggingConfig{
			Level:       "info",
			MaxFileSize: 100,
			MaxAge:      7,
			MaxNumFiles: 10,
			Compress:    true,
		},
		HostConfig: HostConfig{
			Host: "0.0.0.0",
			Port: 8080,
		},
	}

	path := filepath.Join(baseDir, "config")

	configPathVar := fmt.Sprintf("%s_CONFIG_PATH", strings.ToUpper(appName))
	configPath := os.Getenv(configPathVar)
	if configPath != "" {
		path = configPath
	} else {
		path = filepath.Join(path, "config.yaml")
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		logger.Error("Error marshalling config", zap.Error(err))
		return nil
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		logger.Error("Error creating config directory", zap.Error(err))
		return nil
	}

	err = os.WriteFile(path, data, 0644)
	if err != nil {
		logger.Error("Error writing config", zap.Error(err))
		return nil
	}
	return c
}

func GetModule(name string) Module {
	if m, ok := modules[name]; ok {
		return m
	}

	return nil
}
