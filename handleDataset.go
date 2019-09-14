package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/teris-io/shortid"
)

func saveDatas(c *gin.Context) ([]*Dataset, error) {
	file, err := c.FormFile("file")
	if err != nil {
		log.Warnf(`saveUploadFile, read upload file error, details: %s\n`, err)
		return nil, err
	}
	// ext
	ext := filepath.Ext(file.Filename)
	lext := strings.ToLower(ext)
	//path := c.Request.URL.Path
	dir := "./uploads"
	switch lext {
	case CSVEXT, GEOJSONEXT, KMLEXT, GPXEXT, ZIPEXT:
		dir = "./uploads"
	case MBTILESEXT:
		dir = "./uploads"
	default:
		return nil, fmt.Errorf("unsupport format")
	}

	id, _ := shortid.Generate()
	name := strings.TrimSuffix(file.Filename, ext)
	dst := filepath.Join(dir, name+"-"+id+lext)
	if err := c.SaveUploadedFile(file, dst); err != nil {
		return nil, fmt.Errorf(`handleDataset, saving uploaded file error, details: %s`, err)
	}

	ds := &Dataset{
		ID:      id,
		Name:    name,
		Tag:     "",
		Format:  lext,
		Path:    dst,
		Size:    file.Size,
		Geotype: "Unknown",
	}

	dts := make([]*Dataset, 0, 10)
	dts = append(dts, ds)

	return dts, nil
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

	for _, dt := range dss {
		//err = dt.Save
		fmt.Printf("%#v", dt)
	}
}
