package main

import (
	"github.com/chenguan1/msmap/mswrap"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

const (
	VERSION = "1.0.0"
)

var (
	db *gorm.DB
)

func main() {

	// mapserver
	mswrap.Start()

	r := setupRouter()

	log.Infoln("start main server.")
	r.Run(":8090")

	log.Info("exit.")

}
