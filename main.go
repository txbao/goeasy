// 示例入口：演示 goeasy/app 最小用法（非库导出）。
package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"github.com/txbao/goeasy/app"
	"github.com/txbao/goeasy/config"
	zresp "github.com/txbao/goeasy/response"
)

func main() {
	cfg := &config.Config{
		AppName: "goeasy-demo",
		Env:     "dev",
		HTTP:    config.HTTP{Host: "0.0.0.0", Port: 8080},
	}
	application := app.New(cfg)
	application.RegisterHTTP(func(engine *gin.Engine) {
		engine.GET("/", func(c *gin.Context) {
			zresp.Success(c, gin.H{"message": "Welcome to goeasy"})
		})
		engine.GET("/health", func(c *gin.Context) {
			zresp.Success(c, gin.H{"status": "healthy"})
		})
	})
	if err := application.Run(); err != nil {
		log.Fatal(err)
	}
}
