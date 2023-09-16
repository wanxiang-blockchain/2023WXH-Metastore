package config

import "github.com/spf13/viper"

var USER_AGENT string

var GConf = Config{}

type Config struct {
	ApiBaseUrl  string `mapstructure:"api-base-url"`
	Port        int    `mapstructure:"port"`
	ServerMode  string `mapstructure:"server-mode"`
	LogConfig   `mapstructure:"log"`
	DbConfig    `mapstructure:"db"`
	RedisConfig `mapstructure:"redis"`
}

type DbConfig struct {
	Endpoint string `mapstructure:"endpoint"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DbName   string `mapstructure:"db-name"`
}

type RedisConfig struct {
	Address  string `mapstructure:"address"`
	Password string `mapstructure:"password"`
	MaxAge   int    `mapstructure:"maxAge"`
}

type LogConfig struct {
	Dir          string `mapstructure:"dir"`
	Prefix       string `mapstructure:"prefix"`
	Suffix       string `mapstructure:"suffix"`
	Level        string `mapstructure:"level"` // log level
	MaxAge       int    `mapstructure:"maxAge"`
	RotationTime int    `mapstructure:"rotationTime"`
	Development  bool   `mapstructure:"development"`
}

func InitConfig(configPath string) error {
	viper.SetConfigFile(configPath)
	err := viper.ReadInConfig()
	if err != nil {
		return err
	}
	return viper.Unmarshal(&GConf)
}
