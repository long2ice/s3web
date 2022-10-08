package main

import (
	"github.com/NYTimes/gziphandler"
	"net/http"
	"time"

	"github.com/long2ice/s3web/config"
	log "github.com/sirupsen/logrus"
)

func WithLogging(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		uri := r.RequestURI
		method := r.Method
		h.ServeHTTP(w, r)
		duration := time.Since(start)
		log.WithFields(log.Fields{
			"uri":      uri,
			"method":   method,
			"duration": duration,
		}).Info(r.Host)
	})
}

func main() {
	var handler http.Handler
	handler = NewS3Handler()
	if config.ServerConfig.Gzip {
		log.Info("gzip enabled")
		handler = gziphandler.GzipHandler(handler)
	}
	http.Handle("/", WithLogging(handler))
	listen := config.ServerConfig.Listen
	log.Infof("Started listening on %s", listen)
	log.Fatalln(http.ListenAndServe(listen, nil))
}
