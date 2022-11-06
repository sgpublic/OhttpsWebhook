package config

import (
	"OhttpsWebhook/src/module"
	"OhttpsWebhook/src/util"
	"flag"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"os"
	"path"
	"time"
)

var (
	_ConfigPath string
	_Debug      bool
)

func Setup() {
	_SetupFlag()
	_SetupConfig()
}

func _SetupFlag() {
	flag.StringVar(&_ConfigPath, "c", "./config.yaml", "Set the config file (*.yaml) to use.")
	flag.BoolVar(&_Debug, "d", false, "Enable debug mode.")
}

func _SetupConfig() {
	conf := _ReadConfig()
	module.DoWebhook(conf.Hook.Path, conf.Hook.Port)

	logPath := conf.Config.Logging.Path
	if logPath == "" {
		logPath = "./log/"
	}
	rotateOptions := []rotatelogs.Option{
		rotatelogs.WithRotationTime(time.Hour * 24),
	}
	rotateOptions = append(rotateOptions, rotatelogs.WithMaxAge(time.Duration(conf.Config.Logging.Aging)))
	w, err := rotatelogs.New(path.Join(logPath, "%Y-%m-%d.log"), rotateOptions...)
	if err != nil {
		log.Errorf("rotatelogs init err: %v", err)
	} else {
		log.AddHook(util.NewLocalHook(w, _Debug))
	}
}

func GetTarget(domain string) (Target, error) {
	conf := _ReadConfig()
	for _, target := range conf.Targets {
		if target.Domain == domain {
			return target, nil
		}
	}
	return Target{}, &IdNotDefineError{domain: domain}
}

func GetWebhookKey() string {
	return _ReadConfig().Config.Key
}

func _ReadConfig() _ConfigRoot {
	conf := new(_ConfigRoot)
	yamlFile, err := os.ReadFile(_ConfigPath)
	if err != nil {
		log.Fatalf("Config file \"%s\" not found, please create one first!", _ConfigPath)
	}
	err = yaml.Unmarshal(yamlFile, conf)
	if err != nil {
		log.Fatalf("Config file read failed: %v", err)
	}
	return *conf
}

type Target struct {
	Domain         string `yaml:"domain"`
	CertKey        string `yaml:"cert-key"`
	FullchainCerts string `yaml:"fullchain-certs"`
}

type _ConfigRoot struct {
	Hook struct {
		Path string `yaml:"path"`
		Port int    `yaml:"port"`
	} `yaml:"hook"`

	Config struct {
		Key     string `yaml:"key"`
		Logging struct {
			Path  string `yaml:"path"`
			Aging int64  `yaml:"aging"`
		} `yaml:"logging"`
	}

	Targets []Target `yaml:"targets"`
}

type IdNotDefineError struct {
	domain string
}

func (e *IdNotDefineError) Error() string {
	return "Domain of \"" + e.domain + "\" not defined!"
}
