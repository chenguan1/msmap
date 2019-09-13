package mswrap

import (
	"net/http"
	"net/http/cgi"

	log "github.com/sirupsen/logrus"
)

func main() {

	http.HandleFunc("/ms", func(w http.ResponseWriter, r *http.Request) {
		handler := new(cgi.Handler)
		handler.Path = "D:/ms4w/Apache/cgi-bin/mapserv"
		handler.Env = append(handler.Env, "GDAL_DATA=d:/ms4w/gdaldata")
		handler.Env = append(handler.Env, "GDAL_DRIVER_PATH=d:/ms4w/gdalplugins")
		handler.Env = append(handler.Env, "PROJ_LIB=d:/ms4w/proj/nad")

		log.Println(r.RemoteAddr, r.RequestURI)

		handler.ServeHTTP(w, r)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handler := new(cgi.Handler)
		handler.Path = "D:/ms4w/Apache/cgi-bin/mapcache.fcgi.exe"
		handler.Env = append(handler.Env, "MAPCACHE_CONFIG_FILE=D:/ms4w/apps/mapcache/mapcache.xml")

		log.Println(r.RemoteAddr, r.RequestURI)

		handler.ServeHTTP(w, r)
	})

	log.Infoln("Starting MapServer...")
	log.Fatalln(http.ListenAndServe(":8035", nil))

}
