package main

import (
	"fmt"
	"github.com/chenguan1/msmap/mswrap"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/teris-io/shortid"
)

//valSizeShp valid shapefile, return 0 is invalid
func valSizeShp(shp string) int64 {
	ext := filepath.Ext(shp)
	if strings.Compare(SHPEXT, ext) != 0 {
		return 0
	}
	info, err := os.Stat(shp)
	if err != nil {
		return 0
	}
	total := info.Size()

	pathname := strings.TrimSuffix(shp, ext)
	info, err = os.Stat(pathname + ".dbf")
	if err != nil {
		return 0
	}
	total += info.Size()

	info, err = os.Stat(pathname + ".shx")
	if err != nil {
		return 0
	}
	total += info.Size()

	info, err = os.Stat(pathname + ".prj")
	if err != nil {
		return 0
	}
	total += info.Size()

	return total
}

func getDatafiles(dir string) (map[string]int64, error) {
	files := make(map[string]int64)
	itmes, err := ioutil.ReadDir(dir)
	if err != nil {
		return files, err
	}
	for _, item := range itmes {
		name := item.Name()
		if item.IsDir() {
			subfiles, _ := getDatafiles(filepath.Join(dir, name))
			for k, v := range subfiles {
				files[k] = v
			}
		}
		ext := filepath.Ext(name)
		//处理zip内部数据文件
		switch ext {
		case CSVEXT, GEOJSONEXT, KMLEXT, GPXEXT:
			files[filepath.Join(dir, name)] = item.Size()
		case SHPEXT:
			shp := filepath.Join(dir, name)
			size := valSizeShp(shp)
			if size > 0 {
				files[shp] = size
			}
		default:
		}
	}
	return files, nil
}

func loadZipData(zipfile string) ([]*Dataset, error) {
	var dss []*Dataset
	wd, _ := os.Getwd()
	tmpdir := strings.TrimSuffix(zipfile, filepath.Ext(zipfile))
	err := UnZipToDir(zipfile, tmpdir)
	if err != nil {
		return nil, err
	}
	files, err := getDatafiles(tmpdir)
	if err != nil {
		return nil, err
	}
	for file, size := range files {
		subase, err := filepath.Rel(tmpdir, file)
		if err != nil {
			subase = filepath.Base(file)
		}
		wdpath, err := filepath.Rel(wd, file)
		if err != nil {
			return nil, err
		}
		ext := filepath.Ext(file)
		subname := filepath.ToSlash(subase)
		subname = strings.Replace(subname, "/", "_", -1)
		subname = strings.TrimSuffix(subname, ext)
		subid, _ := shortid.Generate()
		subdt := &Dataset{
			ID:        subid,
			Name:      subname,
			Tag:       "",
			Format:    strings.ToLower(ext),
			Path:      wdpath,
			Size:      size,
			Geotype:   "Unknown",
			CreatedAt: time.Time{},
			UpdatedAt: time.Time{},
		}
		dss = append(dss, subdt)
	}

	return dss, nil
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
		dss2, err := loadZipData(fulpath)
		if err != nil {
			return nil, fmt.Errorf(`handleDataset, unzip file error, details: %s`, err)
		}
		dss = append(dss, dss2...)
	}

	// 尝试读取信息，获取它的Geotype

	return dss, nil
}

////////////////////////////////////////////////////////////////
////////////////////////// -handle- ////////////////////////////

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

	//log.Info("begin load datasource")

	for _,dt := range dss{
		err := dt.LoadFrom()
		if err != nil{
			log.Errorf("load from failed, error: %v",err)
		}
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

	// 更新mapfile
	mf := Mapfile{}
	mf.layers = append(mf.layers, dt)
	wd, _ := os.Getwd()
	mapfile := filepath.Join(wd, "data/mapfiles", dt.ID + ".map")
	err = mf.Generate(dt.Name, mapfile)
	if err != nil {
		log.Errorf("generate mapfile failed, error: %v", err)
		res.FailMsg(c, "geneate mapfile failed.")
		return
	}
	dt.Mapfile = mapfile

	//log.Infof("dataset: %#v",dt)

	// 保存到数据库
	if err = dt.Save(); err != nil {
		log.Warnf("create database failed, err: %s", err)
		res.FailErr(c, err)
		return
	}

	res.DoneData(c, dt)
}

// wms 预览数据集
func wmsDataset(c *gin.Context){
	res := NewRes()
	id := c.Param("id")
	dt := &Dataset{}
	err := db.Where("id = ?", id).First(dt).Error
	if err != nil {
		res.Fail(c,4046)
		return
	}

	// path
	c.Request.URL.Path = mswrap.PathMapServ

	// rawquery
	mpf := filepath.ToSlash(dt.Mapfile)
	addParam := "map="+mpf
	if c.Request.URL.RawQuery != ""{
		c.Request.URL.RawQuery = c.Request.URL.RawQuery +  "&"
	}
	c.Request.URL.RawQuery = c.Request.URL.RawQuery +  addParam

	mswrap.ProxyMapServ.ServeHTTP(c.Writer, c.Request)
}

// wms 预览数据集
func xyzDataset(c *gin.Context){
	res := NewRes()
	id := c.Param("id")
	dt := &Dataset{}
	err := db.Where("id = ?", id).First(dt).Error
	if err != nil {
		res.Fail(c,4046)
		return
	}

	// path
	c.Request.URL.Path = mswrap.PathMapServ

	// rawquery
	mpf := filepath.ToSlash(dt.Mapfile)
	addParam := "map="+mpf
	if c.Request.URL.RawQuery != ""{
		c.Request.URL.RawQuery = c.Request.URL.RawQuery +  "&"
	}
	c.Request.URL.RawQuery = c.Request.URL.RawQuery +  addParam

	mswrap.ProxyMapServ.ServeHTTP(c.Writer, c.Request)
}

