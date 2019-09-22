package main

import (
	"fmt"

	"github.com/chenguan1/msmap/mswrap"
	"github.com/jinzhu/gorm"
	//_ "github.com/mattn/go-sqlite3"
	_ "github.com/shaxbee/go-spatialite" // Spatialite SQL Driver for Golang
	log "github.com/sirupsen/logrus"
)

const (
	VERSION = "1.0.0"
)

var (
	db *gorm.DB
)

// 初始化数据库
func initDb() (*gorm.DB, error) {
	var conn string
	dbtype := "sqlite3"
	switch dbtype {
	case "sqlite3":
		conn = "data/msmap.db"
	default:
		return nil, fmt.Errorf("unknown database")
	}

	db, err := gorm.Open(dbtype, conn)
	if err != nil {
		return nil, fmt.Errorf("init gorm db error, error: %s\n", err)
	}

	log.Info("init gorm db succcess.")

	// 自动构建
	/*gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
		return "ms_" + defaultTableName
	}*/
	db.AutoMigrate(&Dataset{}, &Mapset{})
	return db, nil
}

func main() {
	test()
	return

	var err error
	db, err = initDb()
	if err != nil {
		log.Fatalf("init db error, error: %s", err)
	}
	defer db.Close()

	// mapserver
	mswrap.Start()

	r := setupRouter()

	log.Infoln("start main server.")
	r.Run(":8090")

	log.Info("exit.")
}
