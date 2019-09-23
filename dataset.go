package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/axgle/mahonia"
	"github.com/jinzhu/gorm"
	"github.com/jonas-p/go-shp"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
	log "github.com/sirupsen/logrus"
	//"github.com/tingold/orb"
	"golang.org/x/text/encoding/simplifiedchinese"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

//BUFSIZE 16M
const (
	BUFSIZE   int64 = 1 << 24
	PREROWNUM       = 7
)

// Field 字段
type Field struct {
	Name  string    `json:"name"`
	Alias string    `json:"alias"`
	Type  FieldType `json:"type"`
	Index string    `json:"index"`
}

type BBox struct {
	MinX float64    `json:"minx"`
	MinY float64    `json:"miny"`
	MaxX float64    `json:"maxx"`
	MaxY float64    `json:"maxy"`
}

// Dataset 数据集定义结构
type Dataset struct {
	ID        string          `json:"id" gorm:"primary_key"` //字段列表
	Name      string          `json:"name"`                  //字段列表// 数据集名称,现用于更方便的ID
	Tag       string          `json:"tag"`
	Path      string          `json:"path"`
	Format    string          `json:"format"`
	Encoding  string          `json:"encoding"`
	Size      int64           `json:"size"`
	Total     int             `json:"total"`
	BBox      BBox            `json:"bbox"` //gorm:"type:json"
	Crs       string          `json:"crs"` //WGS84,CGCS2000,GCJ02,BD09
	//Rows      [][]string      `json:"rows" gorm:"-"`
	Rows      [][]string      `json:"-" gorm:"-"`
	Geotype   GeoType         `json:"geotype"`
	//Fields    json.RawMessage `json:"fields" gorm:"type:json"` //字段列表
	Fields    json.RawMessage `json:"-" gorm:"-"`
	Mapfile   string          `json:"mapfile"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

func (dt *Dataset) AbsPath() string {
	wd, _ := os.Getwd()
	abspath := filepath.Join(wd,dt.Path)
	return abspath
}

// Save 更新/创建数据（集）
func (dt *Dataset) Save() error {
	tmp := &Dataset{}
	err := db.Where("id = ?", dt.ID).First(tmp).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			dt.CreatedAt = time.Time{}
			err = db.Create(dt).Error
			if err != nil {
				return err
			}
		} else {
			return err
		}

	}
	err = db.Model(&Dataset{}).Update(dt).Error
	return err
}

// 载入数据
func (dt *Dataset) LoadFrom() error {
	switch dt.Format {
	case CSVEXT:
		return dt.LoadFromCSV()
	case GEOJSONEXT:
		return dt.LoadFromJson()
	case SHPEXT:
		return dt.LoadFromShp()
	}
	return nil
}

func likelyEncoding(file string) string {
	stat, err := os.Stat(file)
	if err != nil {
		return ""
	}
	bufsize := BUFSIZE
	if stat.Size() < bufsize {
		bufsize = stat.Size()
	}
	r, err := os.Open(file)
	if err != nil {
		return ""
	}
	defer r.Close()
	buf := make([]byte, bufsize)
	rn, err := r.Read(buf)
	if err != nil {
		return ""
	}
	return Mostlike(buf[:rn])
}

func csvReader(r io.Reader, encoding string) (*csv.Reader, error) {
	switch encoding {
	case "gbk", "big5", "gb18030":
		decoder := mahonia.NewDecoder(encoding)
		if decoder == nil {
			return csv.NewReader(r), fmt.Errorf(`create %s encoder error`, encoding)
		}
		dreader := decoder.NewReader(r)
		return csv.NewReader(dreader), nil
	default:
		return csv.NewReader(r), nil
	}
}


func (dt *Dataset) LoadFromCSV() error {
	if dt.Encoding == "" {
		dt.Encoding = likelyEncoding(dt.AbsPath())
	}
	file, err := os.Open(dt.AbsPath())
	if err != nil {
		return err
	}
	defer file.Close()
	reader, err := csvReader(file, dt.Encoding)
	if err != nil {
		return err
	}
	headers, err := reader.Read()
	if err != nil {
		return err
	}
	var records [][]string
	var rowNum, perNum int
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}
		if perNum < PREROWNUM {
			records = append(records, row)
			perNum++
		}
		rowNum++
	}

	findType := func(arr []string) FieldType {
		var hasFloats, hasInts, hasBools, hasStrings bool
		for _, str := range arr{
			if str == "" {
				continue
			}
			if _, err := strconv.Atoi(str); err == nil{
				hasInts = true
				continue
			}
			if _, err := strconv.ParseFloat(str,64); err == nil{
				hasFloats = true
				continue
			}
			if str == "true" || str == "false"{
				hasBools = true
				continue
			}
			hasStrings = true
		}
		switch {
		case hasStrings:
			return String
		case hasBools:
			return Bool
		case hasFloats:
			return Float
		case hasInts:
			return Int
		default: //all null or string data
			return String
		}
	}

types := make([]FieldType, len(headers))
	for i := range headers {
		col := make([]string, len(records))
		for j := 0; j < len(records); j++ {
			col[j] = records[j][i]
		}
		types[i] = findType(col)
	}

	var fields []Field
	for i, name := range headers {
		fields = append(fields, Field{
			Name: name,
			Type: types[i]})
	}

	getColumn := func(cols []string, names []string) string {
		for _, c := range cols {
			for _, n := range names {
				if c == strings.ToLower(n) {
					return n
				}
			}
		}
		return ""
	}

	detechColumn := func(min float64, max float64) string {
		for i, name := range headers {
			num := 0
			for _, row := range records {
				f, err := strconv.ParseFloat(row[i], 64)
				if err != nil || f < min || f > max {
					break
				}
				num++
			}
			if num == len(records) {
				return name
			}
		}
		return ""
	}

	xcols := []string{"x", "lon", "longitude", "经度"}
	x := getColumn(xcols, headers)
	if x == "" {
		//x = detechColumn(73, 135) // 中国范围
		x = detechColumn(-180, 180) // 中国范围
	}
	ycols := []string{"y", "lat", "latitude", "纬度"}
	y := getColumn(ycols, headers)
	if y == "" {
		//y = detechColumn(18, 54)
		y = detechColumn(-90, 90)
	}

	// 判断
	if x == "" || y == "" {
		return fmt.Errorf("cannot detect axis column.")
	}

	// bbox
	// 查找
	findColIndex := func(name string, names []string) int {
		for i, n := range names {
			if n == name {
				return i
			}
		}
		return -1
	}

	// 最大最小值
	findMinMax := func(idx int) (float64, float64, error) {
		var min, max float64
		for r, row := range records {
			f, err := strconv.ParseFloat(row[idx], 64)
			if err != nil {
				return 0, 0, fmt.Errorf("Parse position falied. error: %v",err)
			}
			if r == 0 {
				min = f
				max = f
			}else{
				if min > f{
					min = f
				}
				if max < f{
					max = f
				}
			}
		}
		return min, max, nil
	}

	idx_x := findColIndex(x, headers)
	idx_y := findColIndex(y, headers)

	var bbox [4]float64
	bbox[0], bbox[2], err = findMinMax(idx_x)
	if err != nil {
		return err
	}
	bbox[1], bbox[3], err = findMinMax(idx_y)
	if err != nil {
		return err
	}
	dt.BBox = BBox{
		MinX:bbox[0],
		MinY:bbox[1],
		MaxX:bbox[2],
		MaxY:bbox[3],
	}

	dt.Format = CSVEXT
	dt.Total = rowNum
	if x != "" && y != "" {
		dt.Geotype = GeoType(x + "," + y)
	}
	dt.Crs = "WGS84"
	dt.Rows = records
	flds, err := json.Marshal(fields)
	if err == nil {
		dt.Fields = flds
	}

	return nil
}


func jsonDecoder(r io.Reader, encoding string) (*json.Decoder, error) {
	switch encoding {
	case "gbk", "big5", "gb18030": //buf reader convert
		mdec := mahonia.NewDecoder(encoding)
		if mdec == nil {
			return json.NewDecoder(r), fmt.Errorf(`create %s encoder error`, encoding)
		}
		mdreader := mdec.NewReader(r)
		return json.NewDecoder(mdreader), nil
	default:
		return json.NewDecoder(r), nil
	}
}

//movetoFeatures move decoder to features
func movetoFeatures(decoder *json.Decoder) error {
	_, err := decoder.Token()
	if err != nil {
		return err
	}
out:
	for {
		t, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		switch v := t.(type) {
		case string:
			if v == "features" {
				t, err := decoder.Token()
				if err != nil {
					return err
				}
				d, ok := t.(json.Delim)
				if ok {
					ds := d.String()
					if ds == "[" {
						break out
					}
				}
			}
		}
	}
	return nil
}

func mergeBBox(box1 orb.Bound, box2 orb.Bound) orb.Bound{
	box := box1
	if box.Min[0] > box2.Min[0] {
		box.Min[0] = box2.Min[0]
	}
	if box.Min[1] > box2.Min[1] {
		box.Min[1] = box2.Min[1]
	}
	if box.Max[0] < box2.Max[0] {
		box.Max[0] = box2.Max[0]
	}
	if box.Max[1] < box2.Max[1] {
		box.Max[1] = box2.Max[1]
	}
	return box
}

func (dt *Dataset) LoadFromJson() error {
	if dt.Encoding == "" {
		dt.Encoding = likelyEncoding(dt.AbsPath())
	}
	file, err := os.Open(dt.AbsPath())
	if err != nil {
		return err
	}
	defer file.Close()

	dec, err := jsonDecoder(file,dt.Encoding)
	if err != nil{
		return err
	}

	s := time.Now()
	err = movetoFeatures(dec)
	if err != nil{
		return err
	}

	prepAttrRow := func(fields []Field, props geojson.Properties) []string {
		var row []string
		for _, field := range fields {
			var s string
			v, ok := props[field.Name]
			if ok {
				switch v.(type) {
				case bool:
					val, ok := v.(bool) // Alt. non panicking version
					if ok {
						s = strconv.FormatBool(val)
					} else {
						s = "null"
					}
				case float64:
					val, ok := v.(float64) // Alt. non panicking version
					if ok {
						s = strconv.FormatFloat(val, 'g', -1, 64)
					} else {
						s = "null"
					}
				case map[string]interface{}, []interface{}:
					buf, err := json.Marshal(v)
					if err == nil {
						s = string(buf)
					}
				default: //string/map[string]interface{}/[]interface{}/nil->对象/数组都作string处理
					if v == nil {
						s = ""
					} else {
						s, _ = v.(string)
					}
				}
			}
			row = append(row, s)
		}
		return row
	}

	var rows [][]string
	var rowNum, preNum int
	ft := &geojson.Feature{}
	err = dec.Decode(ft)
	if err != nil {
		log.Errorf(`geojson data format error, details:%s`, err)
		return err
	}

	// box
	bbox := ft.Geometry.Bound()
	//log.Infof("bbox: %v",ft.Geometry.Bound())

	rowNum++
	preNum++
	geoType := ft.Geometry.GeoJSONType()
	var fields []Field
	for k, v := range ft.Properties {
		var t FieldType
		switch v.(type) {
		case bool:
			t = Bool //or 'timestamp with time zone'
		case float64:
			t = Float
		default: //string/map[string]interface{}/[]interface{}/nil->对象/数组都作string处理
			t = String
		}
		fields = append(fields, Field{
			Name: k,
			Type: t,
		})
	}
	row := prepAttrRow(fields, ft.Properties)
	rows = append(rows, row)

	for dec.More() {
		if preNum < PREROWNUM {
			ft := &geojson.Feature{}
			err := dec.Decode(ft)
			if err != nil {
				log.Errorf(`geojson data format error, details:%s`, err)
				continue
			}
			bbox = mergeBBox(bbox, ft.Geometry.Bound())
			rows = append(rows, prepAttrRow(fields, ft.Properties))
			preNum++
		} else {
			var ft struct {
				Type string `json:"type"`
			}
			err := dec.Decode(&ft)
			if err != nil {
				log.Errorf(`Decode error, details:%s`, err)
				continue
			}
		}
		rowNum++
	}
	fmt.Printf("total features %d, takes: %v\n", rowNum, time.Since(s))

	dt.BBox = BBox{
		MinX:bbox.Min[0],
		MinY:bbox.Min[1],
		MaxX:bbox.Max[0],
		MaxY:bbox.Max[1],
	}

	dt.Format = GEOJSONEXT
	dt.Total = rowNum
	dt.Geotype = GeoType(geoType)
	dt.Crs = "WGS84"
	dt.Rows = rows
	flds, err := json.Marshal(fields)
	if err == nil {
		dt.Fields = flds
	}


	return nil
}

func (dt *Dataset) LoadFromShp() error {
	size := valSizeShp(dt.AbsPath())
	if size == 0 {
		return fmt.Errorf("invalid shapefiles")
	}

	shape, err := shp.Open(dt.AbsPath())
	if err != nil {
		return err
	}
	defer shape.Close()

	bbox := shape.BBox()
	//log.Infof("geotype: %v, box: %v", gtype, box)

	shpfield := shape.Fields()
	total := shape.AttributeCount()

	var fields []Field
	for _,v := range shpfield {
		var t FieldType
		switch v.Fieldtype {
		case 'C':
			t = String
		case 'N':
			t = Int
		case 'F':
			t = Float
		case 'D':
			t = Date
		}
		fn := v.String()
		ns, err := simplifiedchinese.GB18030.NewDecoder().String(fn)
		if err == nil {
			fn = ns
		}
		fields = append(fields, Field{
			Name: fn,
			Type: t,
		})
	}

	rowstxt := ""
	var rows [][]string
	preRowNum := 0
	for shape.Next(){
		if preRowNum > PREROWNUM {
			break
		}
		var row []string
		for k := range fields {
			v := shape.Attribute(k)
			row = append(row, v)
			rowstxt += v
		}
		rows = append(rows, row)
		preRowNum++
	}

	if dt.Encoding == "" {
		dt.Encoding = Mostlike([]byte(rowstxt))
	}
	var mdec mahonia.Decoder
	switch dt.Encoding {
	case "gbk", "big5", "gb18030":
		mdec = mahonia.NewDecoder(dt.Encoding)
		if mdec != nil {
			var records [][]string
			for _, row := range rows {
				var record []string
				for _, v := range row {
					record = append(record, mdec.ConvertString(v))
				}
				records = append(records, record)
			}
			rows = records
		}
	}

	var geoType string
	switch shape.GeometryType {
	case 1: //POINT
		geoType = "Point"
	case 3: //POLYLINE
		geoType = "LineString"
	case 5: //POLYGON
		geoType = "MultiPolygon"
	case 8: //MULTIPOINT
		geoType = "MultiPoint"
	}

	dt.BBox = BBox{
		MinX:bbox.MinX,
		MinY:bbox.MinY,
		MaxX:bbox.MaxX,
		MaxY:bbox.MaxY,
	}

	dt.Format = SHPEXT
	dt.Size = size
	dt.Total = total
	dt.Geotype = GeoType(geoType)
	dt.Crs = "WGS84"
	dt.Rows = rows
	jfs, err := json.Marshal(fields)
	if err == nil {
		dt.Fields = jfs
	} else {
		log.Error(err)
	}
	return nil
}
