package main

import (
	"encoding/json"
	"io/ioutil"
)

// Config is the configuration object used by raspberry to set up various parameters.
type Config struct {
	ListenPath     string `json:"listen_path"`
	ListenPort     int    `json:"listen_port"`
	TargetUrl      string `json:"target_url"`
	Secret         string `json:"secret"`
	TemplatePath   string `json:"template_path"`
	AuthHeaderName string `json:"auth_header_name"`
	Storage        struct {
		Type     string `json:"type"`
		Host     string `json:"host"`
		Port     int    `json:"port"`
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"storage"`
	ExcludePaths []string `json:"exclude_paths"`
}

// WriteDefaultConf will create a default configuration file and set the storage type to "memory"
func WriteDefaultConf(configStruct *Config) {
	configStruct.ListenPath = "/gateway"
	configStruct.ListenPort = 8080
	configStruct.TargetUrl = "http://localhost:8080/api"
	configStruct.Secret = "352d20ee67be67f6340b4c0605b044b7"
	configStruct.TemplatePath = "templates"
	configStruct.Storage.Type = "momery"
	configStruct.Storage.Host = "localhsot"
	configStruct.Storage.Port = 6379
	configStruct.Storage.Username = "user"
	configStruct.Storage.Password = "password"
	newConfig, err := json.Marshal(configStruct)
	if err != nil {
		log.Error("Problem marshalling default configuration")
		log.Error(err)
	} else {
		ioutil.WriteFile("raspberry.conf", newConfig, 0644)
	}
}

// LoadConfig will load the configuration file from filePath, if it can't open
// the file for reading, it assumes there is no configuration file and will try to create
// one on the default path (raspberry.conf in the local directory)
func loadConfig(filePath string, configStruct *Config) {
	configuration, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Error("Couldn't load configuration file")
		log.Error(err)
		log.Info("Writing a default file to ./raspberry.conf")

		WriteDefaultConf(configStruct)

		log.Info("Loading default configuration...")
		loadConfig("raspberry.conf", configStruct)
	} else {
		err := json.Unmarshal(configuration, &configStruct)
		if err != nil {
			log.Error("Couldn't unmarshal configuration")
			log.Error(err)
		}
	}
}
