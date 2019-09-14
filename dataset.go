package main

import (
	"time"
)

// Dataset 数据集定义结构
type Dataset struct {
	ID        string    `json:"id" gorm:"primary_key"` //字段列表
	Name      string    `json:"name"`                  //字段列表// 数据集名称,现用于更方便的ID
	Tag       string    `json:"tag"`
	Path      string    `json:"-"`
	Format    string    `json:"format"`
	Size      int64     `json:"size"`
	Geotype   string    `json:"geotype"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Save 更新/创建数据（集）
func (dt *Dataset) Save() error {
	tmp := &Dataset{}
	err := db.Where("id = ?", dt.ID).First(tmp).Error
	return err
}
