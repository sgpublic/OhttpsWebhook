package config

import (
	"OhttpsWebhook/src/util"
	"flag"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"
)

var (
	_ConfigPath string
	_Debug      bool
	_Service    bool
)

func Setup() {
	_SetupFlag()
	_SetupConfig()
	if _Service {
		Daemon()
	}
}

func _SetupFlag() {
	flag.StringVar(&_ConfigPath, "c", "./config.yaml", "Set the config file (*.yaml) to use.")
	flag.BoolVar(&_Debug, "d", false, "Enable debug mode.")
	flag.BoolVar(&_Service, "s", false, "Run on daemon mode")
}

func _SetupConfig() {
	conf := _ReadConfig()

	if _Debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
	logPath := conf.Config.Logging.Path
	if logPath == "" {
		logPath = "./log/"
	}
	rotateOptions := []rotatelogs.Option{
		rotatelogs.WithRotationTime(time.Hour * 24),
	}
	aging := conf.Config.Logging.Aging
	if aging == 0 {
		aging = 259200
	}
	rotateOptions = append(rotateOptions, rotatelogs.WithMaxAge(time.Duration(aging)))
	w, err := rotatelogs.New(path.Join(logPath, "%Y-%m-%d.log"), rotateOptions...)
	if err != nil {
		log.Errorf("rotatelogs init err: %v", err)
	} else {
		log.AddHook(util.NewLocalHook(w, _Debug))
	}
}

func Daemon() {
	args := os.Args[1:]
	execArgs := make([]string, 0)
	l := len(args)
	for i := 0; i < l; i++ {
		if strings.Index(args[i], "-s") == 0 {
			continue
		}
		execArgs = append(execArgs, args[i])
	}
	ex, _ := os.Executable()
	p, _ := filepath.Abs(ex)
	proc := exec.Command(p, execArgs...)
	err := proc.Start()
	if err != nil {
		panic(err)
	}
	log.Infof("[PID] %d", proc.Process.Pid)
	os.Exit(0)
}

func GetServiceTarget() (string, string) {
	conf := _ReadConfig()
	return conf.Hook.Path, conf.Hook.Listen
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
		log.Fatalf("Config file '%s' not found, please create one first!", _ConfigPath)
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
		Path   string `yaml:"path"`
		Listen string `yaml:"listen"`
	} `yaml:"hook"`

	Config struct {
		Key     string `yaml:"key"`
		Logging struct {
			Path  string `yaml:"path"`
			Aging int64  `yaml:"aging"`
		} `yaml:"logging"`
	} `yaml:"config"`

	Targets []Target `yaml:"targets"`
}

type IdNotDefineError struct {
	domain string
}

func (e *IdNotDefineError) Error() string {
	return "Domain of \"" + e.domain + "\" not defined!"
}
