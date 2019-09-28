package main

// CRS coordinate reference system
type CRS string

// Supported CRSs
const (
	WGS84    CRS = "WGS84"
	CGCS2000     = "CGCS2000"
	GCJ02        = "GCJ02"
	BD09         = "BD09"
)

//CRSs 支持的坐标系
var CRSs = []CRS{WGS84, CGCS2000, GCJ02, BD09}

// Constants representing TileFormat types
const (
	// ZIPEXT     DataFormat = ".zip"
	ZIPEXT     = ".zip"
	CSVEXT     = ".csv"
	SHPEXT     = ".shp"
	KMLEXT     = ".kml"
	GPXEXT     = ".gpx"
	GEOJSONEXT = ".geojson"
	MBTILESEXT = ".mbtiles"
)

// FieldType is a convenience alias that can be used for a more type safe way of
// reason and use Series types.
type FieldType string

// Supported Series Types
const (
	String      FieldType = "string"
	Bool                  = "bool"
	Int                   = "int"
	Float                 = "float"
	Date                  = "date"
	StringArray           = "string_array"
	Geojson               = "geojson"
)

//GeoType 几何类型
type GeoType string

// A list of the datasets types that are currently supported.
const (
	Point           GeoType = "Point"
	MultiPoint              = "MultiPoint"
	LineString              = "LineString"
	MultiLineString         = "MultiLineString"
	Polygon                 = "Polygon"
	MultiPolygon            = "MultiPolygon"
	Attribute               = "Attribute" //属性数据表,non-spatial
)

//GeoTypes 支持的字段类型
var GeoTypes = []GeoType{Point, MultiPoint, LineString, MultiLineString, Polygon, MultiPolygon, Attribute}
