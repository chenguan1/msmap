package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/teris-io/shortid"
)

func loadZipFile(path string, dss []*Dataset) error {
	return nil
}

func saveDatas(c *gin.Context) ([]*Dataset, error) {
	// working folder
	var wd string
	var err error
	if wd, err = os.Getwd(); err != nil {
		return nil, fmt.Errorf("get wd faild, error: ", err)
	}

	// file updated
	file, err := c.FormFile("file")
	if err != nil {
		log.Warnf(`saveUploadFile, read upload file error, details: %s`, err)
		return nil, err
	}

	// ext
	ext := filepath.Ext(file.Filename)
	lext := strings.ToLower(ext)

	// subdir
	subdir := "data/uploads"
	switch lext {
	case CSVEXT, GEOJSONEXT, KMLEXT, GPXEXT, ZIPEXT:
		subdir = "data/uploads"
	case MBTILESEXT:
		subdir = "data/uploads"
	default:
		return nil, fmt.Errorf("unsupport format")
	}

	dir := filepath.Join(wd, subdir)
	// ensure dir exist
	_, err = os.Stat(dir)
	if os.IsNotExist(err) {
		if err = os.Mkdir(dir, 0755); err != nil {
			return nil, fmt.Errorf("mkdir failed, error: ", err)
		}
	}

	// generate id
	id, _ := shortid.Generate()

	// file name
	name := strings.TrimSuffix(file.Filename, ext)
	name_new := name + "_" + id + lext

	subpath := filepath.Join(subdir, name_new)
	fulpath := filepath.Join(wd, subpath)
	if err := c.SaveUploadedFile(file, fulpath); err != nil {
		return nil, fmt.Errorf(`handleDataset, saving uploaded file error, details: %s`, err)
	}

	// 数据列表
	dss := make([]*Dataset, 0, 128)

	// 压缩包，或非压缩包
	if lext != ".zip" {
		dt := &Dataset{
			ID:      id,
			Name:    name,
			Tag:     "",
			Format:  lext,
			Path:    subpath,
			Size:    file.Size,
			Geotype: "Unknown",
		}
		dss = append(dss, dt)
	} else {
		err = loadZipFile(fulpath, dss)
		if err != nil {
			//
		}
	}

	return dss, nil
}

func uploadData(c *gin.Context) {
	res := NewRes()
	dss, err := saveDatas(c)

	if err != nil {
		log.Warn(err)
		res.FailErr(c, err)
		return
	}

	if dss == nil || len(dss) == 0 {
		res.FailMsg(c, "未找到数据")
	}

	// 上传结果返回前端，前端决定是否导入
	/*for _, dt := range dss {
		err = dt.Save()
		fmt.Printf("%#v", dt)
	}*/

	res.DoneData(c, dss)
}

func listDataset(c *gin.Context) {
	res := NewRes()

	var dss []Dataset
	tdb := db

	kw, y := c.GetQuery("keyword")
	if y {
		tdb = tdb.Where("name ~ ?", kw)
	}
	order, y := c.GetQuery("order")
	if y {
		log.Info(order)
		tdb = tdb.Order(order)
	}
	total := 0
	err := tdb.Model(&Dataset{}).Count(&total).Error
	if err != nil {
		res.Fail(c, 5001)
		return
	}
	start := 0
	rows := 10
	if offset, y := c.GetQuery("start"); y {
		rs, yr := c.GetQuery("rows") //limit count defaut 10
		if yr {
			ri, err := strconv.Atoi(rs)
			if err == nil {
				rows = ri
			}
		}
		start, _ = strconv.Atoi(offset)
		tdb = tdb.Offset(start).Limit(rows)
	}
	err = tdb.Find(&dss).Count(&total).Error
	if err != nil {
		res.Fail(c, 5001)
		return
	}
	res.DoneData(c, gin.H{
		"keyword": kw,
		"order":   order,
		"start":   start,
		"rows":    rows,
		"total":   total,
		"list":    dss,
	})
}

func createDataset(c *gin.Context) {
	res := NewRes()
	dt := Dataset{}
	err := c.Bind(&dt)

	if err != nil {
		log.Warnf("Bind Dataset failed.err")
		res.FailErr(c, err)
		return
	}

	if err = dt.Save(); err != nil {
		log.Warnf("create database failed, err: %s", err)
		res.FailErr(c, err)
		return
	}

	res.Done(c, "ok")

}
