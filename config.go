package gpool

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	log "github.com/sirupsen/logrus"
)

//Config pool config
type Config struct {
	//InitialPoolSize initial pool size. Default: 5
	InitialPoolSize int
	//MinPoolSize min item in pool. Default: 2
	MinPoolSize int
	//MaxPoolSize  max item in pool. Default: 15
	MaxPoolSize int
	//AcquireRetryAttempts retry times when get item Failed. Default: 5
	AcquireRetryAttempts int
	//AcquireIncrement create item count when pool is empty. Default: 5
	AcquireIncrement int
	//TestDuration interval time between check item avaiable.Unit:Millisecond Default: 1000
	TestDuration int
	//TestOnGetItem test avaiable when get item. Default: false
	TestOnGetItem bool
	//Params item initial params
	Params map[string]string
}

//String String
func (config *Config) String() string {
	result := fmt.Sprintf("InitialPoolSize : %d \n MinPoolSize : %d \n MaxPoolSize : %d \n AcquireRetryAttempts : %d \n AcquireIncrement : %d \n TestDuration : %d \n TestOnGetItem : %t \n",
		config.InitialPoolSize,
		config.MinPoolSize,
		config.MaxPoolSize,
		config.AcquireRetryAttempts,
		config.AcquireIncrement,
		config.TestDuration,
		config.TestOnGetItem,
	)
	result = result + "Params:\n"
	for key, value := range config.Params {
		result = result + fmt.Sprintf("\t%s : %s \n", key, value)
	}
	return result
}

//LoadToml load config from toml file
func (config *Config) LoadToml(tomlFilePath string) error {
	log.WithField("tomlFileName", tomlFilePath).Debugf("load config")
	inf, err := os.Stat(tomlFilePath)
	if err != nil {
		log.WithError(err).Error("load toml config ERROR - FILE NOT EXIST ")
		return err
	}
	if !strings.HasSuffix(inf.Name(), ".toml") {
		log.WithFields(log.Fields{
			"need": "*.toml",
			"got":  inf.Name,
		}).Error("load toml config ERROR - FILE TYPE ERROR")
		return errors.New("FILE TYPE ERROR")
	}
	_, err = toml.DecodeFile(tomlFilePath, config)
	return err
}

//DefaultConfig create default config
func DefaultConfig() Config {
	return Config{
		InitialPoolSize:      5,
		MinPoolSize:          2,
		MaxPoolSize:          15,
		AcquireRetryAttempts: 5,
		AcquireIncrement:     5,
		TestDuration:         60000,
		TestOnGetItem:        false,
		Params:               make(map[string]string),
	}
}
