package mswrap

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
)

type MapColor [3]uint8
func (clr *MapColor) String() string{
	return fmt.Sprintf("%v %v %v",clr[0], clr[1], clr[2])
}

type MapBound [4]float64
func (bound *MapBound) String() string{
	return fmt.Sprintf("%v %v %v %v",bound[0], bound[1], bound[2], bound[3])
}
func (bound *MapBound) Union(box2 *MapBound) MapBound{
	bd := *bound
	if bd[0] > box2[0] {
		bd[0] = box2[0]
	}
	if bd[1] > box2[1] {
		bd[1] = box2[1]
	}
	if bd[2] < box2[2] {
		bd[2] = box2[2]
	}
	if bd[3] < box2[3] {
		bd[3] = box2[3]
	}
	return bd
}

func NewMapColor(r, g, b uint8) MapColor{
	return MapColor{r,g,b}
}

func NewMapBound(minx, miny, maxx, maxy float64) MapBound{
	return MapBound{minx, miny, maxx, maxy}
}

type MapLayer struct {
	Name         string
	Data         string
	Geotype      string
	Color        MapColor
	OutlineColor MapColor
}

type MapConfig struct {
	Name    string
	BBox    MapBound
	Mapfile string
	Mshost  string

	Layers  []MapLayer
}

// 生成 mapfile 文件
func (mc *MapConfig) GenerateMapfile(mapfile string) error {
	if mc == nil {
		return  fmt.Errorf("mapfile is nil")
	}
	if mc.Name == "" {
		return fmt.Errorf("mapfile Name is empty")
	}
	if mapfile == "" {
		return fmt.Errorf("mapfile path is empty")
	}

	// \\ -> /
	mapfile = filepath.ToSlash(mapfile)

	// 更新路径
	mc.Mapfile = mapfile
	if mc.Mshost == "" {
		mc.Mshost = UrlMapServ // "http://127.0.0.1:8049/api/v1/ms"
	}

	for i, layer := range mc.Layers{
		if layer.Name == ""{
			mc.Layers[i].Name = mc.Name + "_layer" + strconv.Itoa(i)
		}
		if _, err := os.Stat(layer.Data); os.IsNotExist(err){
			return fmt.Errorf("mapfile layer data is not exist: %s", layer.Data)
		}
		if layer.Geotype == "" {
			return fmt.Errorf("mapfile geotype is empty")
		}

		mc.Layers[i].Data = filepath.ToSlash(layer.Data)
	}

	pt, _ := filepath.Split(mapfile)
	if _,err := os.Stat(pt); os.IsNotExist(err){
		err = os.MkdirAll(pt,0666)
		if err != nil{
			return fmt.Errorf("make dir failed, error: %v", err)
		}
	}

	var b = &strings.Builder{}
	tmpl, err := template.ParseFiles("./static/tmpl/mapfile.tmpl")
	if err != nil {
		return err
	}

	err = tmpl.Execute(b, mc)
	if err != nil{
		return fmt.Errorf("execute template failed, error: %v", err)
	}

	err = ioutil.WriteFile(mapfile, []byte(b.String()), 0666)
	if err != nil{
		return fmt.Errorf("save mapfile failed, error: %v", err)
	}

	//fmt.Println(b.String())

	return nil
}
