package config

import (
	"fmt"
	"github.com/spf13/viper"
	"path/filepath"
)

type Args struct {
	Port int
	User string
	Pwd  string
	Host string
	DB   string
}

func ParseArgs(filename string, path string) Args {
	//viper.AddConfigPath(path)
	//viper.SetConfigName(filename)
	fullPath := filepath.Join(path, filename)
	viper.SetConfigFile(fullPath)
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
	return Args{
		Port: viper.GetInt("server.port"),
		User: viper.GetString("database.user"),
		Pwd:  viper.GetString("database.pwd"),
		Host: viper.GetString("database.host"),
		DB:   viper.GetString("database.db"),
	}
}
