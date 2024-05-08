package utils

import (
	"fmt"
	"github.com/charmbracelet/log"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
	"strings"
	"sync"
)

var (
	onceEnv    sync.Once
	onceConfig sync.Once
)

func Load(configFile string, configStruct any, defaults map[string]interface{}, envKeys map[string]string) error {
	var err error
	onceConfig.Do(func() {
		err = loadConfiguration(configFile, configStruct, defaults, envKeys)
		if err != nil {
			err = fmt.Errorf("error loading configuration: %w", err)
		}
	})

	return err
}

func loadConfiguration(configFile string, configStruct any, defaults map[string]interface{}, envKeys map[string]string) error {
	onceEnv.Do(loadEnv)

	v := viper.New()
	v.SetConfigType("yaml")

	for k, val := range defaults {
		v.SetDefault(k, val)
	}

	// Bind environment variables
	for k, val := range envKeys {
		err := v.BindEnv(k, strings.ToUpper(val))
		if err != nil {
			fmt.Println("Error binding env var", val, err)
		}
	}

	if configFile != "" {
		v.SetConfigFile(configFile)
		if err := v.ReadInConfig(); err != nil {
			return fmt.Errorf("error reading config file: %w", err)
		}
	}

	if err := v.Unmarshal(configStruct); err != nil {
		return fmt.Errorf("error unmarshalling config: %w", err)
	}

	return nil
}

func loadEnv() {
	err := godotenv.Load(".env")

	if err != nil {
		log.Debug("no .env file found")
	}
}
