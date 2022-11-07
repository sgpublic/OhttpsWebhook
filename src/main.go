package launcher

import (
	"OhttpsWebhook/src/config"
	"OhttpsWebhook/src/module"
	"OhttpsWebhook/src/util"
	"github.com/mattn/go-colorable"
	log "github.com/sirupsen/logrus"
)

func Launch() {
	_InitLogrus()
	config.Setup()
	module.Setup()
}

func _InitLogrus() {
	log.SetOutput(colorable.NewColorableStdout())
	log.SetFormatter(util.LogFormat{EnableColor: true})
}
