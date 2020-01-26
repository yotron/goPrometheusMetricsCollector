package common

import (
	"errors"
	. "github.com/yotron/goConfigurableLogger"
	"gopkg.in/yaml.v2"
)

type confSecret struct {
	Passwords map[string]string `yaml:"passwords"`
}

func (conf *confSecret) readSecretPasswordFile(pwFile string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New("Could not read secret file")
		}
		conf = new(confSecret)
	}()
	err = yaml.Unmarshal(ReadYAMLFile(pwFile, conf), &conf)
	if err != nil {
		Error.Println("Error during unmarshalling.")
	}
	return err
}

func (conf *confSecret) getPasswordByEnvExportType() (pw string, err error) {
	Debug.Println("Conf transferred", *conf)
	pw = conf.Passwords[ExportType]
	if pw != "" {
		return pw, nil
	} else {
		return "", errors.New("Password for ExportType not found " + ExportType)
	}
}
