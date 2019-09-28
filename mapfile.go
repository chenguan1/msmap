package main

import (
	"fmt"
	"github.com/chenguan1/msmap/mswrap"
	"strings"
)

type Mapfile struct {
	layers []Dataset
}

// 生成
func (mf *Mapfile) Generate(name string, mapfile string) error {
	mc := mswrap.MapConfig{}
	mc.Name = name
	mc.BBox = mswrap.NewMapBound(-180,-90,180,90)
	mc.Layers = make([]mswrap.MapLayer,0,128)

	for i, layer :=  range mf.layers{
		mcLayer := mswrap.MapLayer{
			Table: "lyr_"+strings.ToLower(layer.ID),
			Name: layer.Name,
			Data: layer.AbsPath(),
			Geotype: string(layer.Geotype),
			Color: mswrap.NewMapColor(255,0,0),
			OutlineColor:mswrap.NewMapColor(0,0,255),
		}

		if strings.Contains(mcLayer.Geotype, "Polygon"){
			mcLayer.Geotype = "Polygon"
		}else if strings.Contains(mcLayer.Geotype, "Line"){
			mcLayer.Geotype = "Line"
		}

		mc.Layers = append(mc.Layers, mcLayer)
		//box2 := mswrap.NewMapBound(layer.BBox.MinX, layer.BBox.MinY, layer.BBox.MaxX, layer.BBox.MaxY)
		box2 := mswrap.NewMapBound(layer.MinX, layer.MinY, layer.MaxX, layer.MaxY)
		if i == 0 {
			mc.BBox = box2
		}else{
			mc.BBox = mc.BBox.Union(&box2)
		}
	}

	err := mc.GenerateMapfile(mapfile)
	if err != nil {
		return fmt.Errorf("Geneate mapfile failed. error: %v",err)
	}

	return nil
}