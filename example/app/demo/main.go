package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/lincaiyong/arg"
	"net/http"
	"os"
)

func main() {
	arg.Parse()
	portStr := arg.KeyValueArg("port", "8989")
	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello World!")
	})
	err := r.Run(fmt.Sprintf("127.0.0.1:%s", portStr))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
