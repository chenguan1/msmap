package main

import (
	"os"
	"testing"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
	"github.com/teris-io/shortid"
)

func TestDatasetAddTake(t *testing.T) {
	dbtype := "sqlite3"
	conn := "msmap_test.db"
	os.Remove(conn)
	var err error
	db, err = gorm.Open(dbtype, conn)
	if err != nil {
		t.Errorf("init gorm db error, error: %s\n", err)
		return
	}

	// 自动构建
	db.AutoMigrate(&Dataset{}, &Mapset{})

	id, _ := shortid.Generate()

	// 创建
	dt := &Dataset{
		ID:        id,
		Name:      id,
		Tag:       "",
		Path:      "d:/test.geojson",
		Format:    ".geojson",
		Size:      12564,
		Geotype:   "line",
		CreatedAt: time.Time{},
		UpdatedAt: time.Time{},
	}

	// 更新
	err = dt.Save()
	if err != nil {
		t.Errorf("insert dataset falied, error: %s\n", err)
		return
	}

	// 取出
	tmp := &Dataset{}
	err = db.Where("id = ?", id).First(tmp).Error
	if err != nil {
		t.Errorf("take one dataset failed, error: %s\n", err)
		return
	}

	if tmp.ID != dt.ID || tmp.Path != dt.Path {
		t.Errorf("data not correct same.")
		t.Errorf("dt=%#v", dt)
		t.Errorf("tm=%#v", tmp)
		return
	}
}
