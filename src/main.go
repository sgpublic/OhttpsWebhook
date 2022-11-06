package launcher

import (
	"OhttpsWebhook/src/config"
	"OhttpsWebhook/src/util"
	"github.com/mattn/go-colorable"
	log "github.com/sirupsen/logrus"
)

func Launch() {
	config.Setup()
	_InitLogrus()
}

func _InitLogrus() {
	log.SetOutput(colorable.NewColorableStdout())
	log.SetFormatter(util.LogFormat{EnableColor: true})
}
