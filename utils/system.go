package utils

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/go-playground/validator.v9"
	"sigs.k8s.io/yaml"

	"go-cygnus/constants"
)

type systemConfig struct {
	SentryConf struct {
		Dsn         string `json:"dsn" validate:"required"`
		Environment string `json:"environment" validate:"required"`
	} `json:"sentry" validate:"required"`
}

var SysConfig systemConfig

func SystemInit() {
	systemYmlFile := fmt.Sprintf("%s/system.yml", constants.ConfigPath)

	var sysConfig systemConfig

	systemContent, readErr := ioutil.ReadFile(systemYmlFile)
	if readErr != nil {
		panic("Read system.yml error, file path:" + systemYmlFile)
	}

	unmarshalErr := yaml.Unmarshal(systemContent, &sysConfig)
	if unmarshalErr != nil {
		panic("Unmarshal system content error.")
	}

	if err := validator.New().Struct(&sysConfig); err != nil {
		panic(fmt.Sprintf("invalid system config: %s", err))
	}

	SysConfig = sysConfig
}
