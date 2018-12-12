package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestDomainLocking(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()

	router.Use(makeDomainLock(
}
