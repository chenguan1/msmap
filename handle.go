package main

import (
	"net/http"
	//"time"

	"github.com/gin-gonic/gin"
)

func hello(c *gin.Context) {
	c.String(http.StatusOK, "Hello There")
}

func ping(c *gin.Context) {
	c.String(http.StatusOK, "Pong")
}
