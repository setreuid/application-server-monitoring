package main

import (
	"context"
	"fmt"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Api struct{}

func (self *Api) Run() {
	router := self.setupRouter()

	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", "", PORT),
		Handler: router,
	}

	go func() {
		if IS_SSL_ENABLE {
			if err := srv.ListenAndServeTLS(SSL_CERT_FILE, SSL_KEY_FILE); err != nil && err != http.ErrServerClosed {
				RUNNING = false
				LogFatal("listen(TLS): %s\n", err)
			}
		} else {
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				RUNNING = false
				LogFatal("listen: %s\n", err)
			}
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	LogInfo("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		RUNNING = false
		LogFatal("Server Shutdown: ", err)
	}
}

func (self *Api) setupRouter() *gin.Engine {
	gin.DisableConsoleColor()
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = ioutil.Discard

	r := gin.Default()
	r.Use(gzip.Gzip(gzip.BestCompression))
	r.Use(CORSMiddleware())

	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	if IS_MASTER {
		r.GET("/status", ctr.Status)

		r.POST("/ping", ctr.Ping)
		r.POST("/hdd", ctr.Hdd)

		if IS_HB_ENABLE {
			r.POST("/hb/:id", hb.Ping)
		}
	}

	return r
}
