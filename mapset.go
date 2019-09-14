package main

import (
	//"encoding/json"
	"time"
	//"github.com/jinzhu/gorm"
)

// Mapset 地图集定义结构
type Mapset struct {
	ID        string    `json:"id" gorm:"primary_key"` //字段列表
	Name      string    `json:"name"`                  //字段列表// 数据集名称,现用于更方便的ID
	Tag       string    `json:"tag"`
	MapFile   string    `json:"-"`
	Version   int       `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
