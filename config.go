package main

import (
	"log"

	"github.com/spf13/viper"
)

type VaultConfiguration struct {
	Address string
}

type Configuration struct {
	Vault VaultConfiguration
}

func initializeConfiguration() {
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")
	viper.AddConfigPath("$HOME/.rv")
	var configuration Configuration

	_ = viper.ReadInConfig()

	err := viper.Unmarshal(&configuration)
	if err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
	}
}

func getConfigValue(val string) interface{} {
	return viper.Get(val)
}

func getConfigValueString(val string) string {
	return viper.GetString(val)
}

func setConfigValue(key string, value interface{}) {
	viper.Set(key, value)
}
