package mswrap

import (
	"net/http"
	"net/http/cgi"
	"net/http/httputil"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// Mapserver Host Info
const HostMapServ = "127.0.0.1:8049"
const UrlMapServ = "http://127.0.0.1:8049/api/v1/ms"
const PathMapServ = "/api/v1/ms"

var ProxyMapServ = httputil.ReverseProxy{
	Director: func(req *http.Request) {
		req.URL.Scheme = "http"
		req.URL.Host = HostMapServ
		req.Host = HostMapServ
	},
}

func HandleMapServ(ctx *gin.Context) {
	ctx.Request.Header.Add("requester-uid", "id")
	ProxyMapServ.ServeHTTP(ctx.Writer, ctx.Request)
}

func Start() {
	http.HandleFunc("/api/v1/ms", func(w http.ResponseWriter, r *http.Request) {
		handler := new(cgi.Handler)
		handler.Path = "D:/ms4w/Apache/cgi-bin/mapserv"
		handler.Env = append(handler.Env, "GDAL_DATA=d:/ms4w/gdaldata")
		handler.Env = append(handler.Env, "GDAL_DRIVER_PATH=d:/ms4w/gdalplugins")
		handler.Env = append(handler.Env, "PROJ_LIB=d:/ms4w/proj/nad")

		log.Println(r.RemoteAddr, r.RequestURI)

		handler.ServeHTTP(w, r)
	})

	// 启动服务
	go func() {
		log.Infoln("start mswraper.")
		log.Fatalln(http.ListenAndServe(":8049", nil))
	}()
}
