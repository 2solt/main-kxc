// main package
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	AUX_URL string `envconfig:"AUX_URL" required:"true"`
	VERSION string `envconfig:"VERSION" required:"true"`
}

type response struct {
	Version    string `json:"version"`
	AuxVersion string `json:"aux_version"`
	Data       any    `json:"data"`
}

type auxResponse struct {
	Version string `json:"version"`
	Data    any    `json:"data"`
}

func callAux(auxURL string, auxPath string) (*http.Response, error) {
	req, _ := http.NewRequest(http.MethodGet, auxURL+auxPath, nil)
	return http.DefaultClient.Do(req)
}

func proxyHandler(c *gin.Context, auxPath string, cfg Config) {
	fmt.Println(cfg.AUX_URL)
	resp, err := callAux(cfg.AUX_URL, auxPath)
	if err != nil {
		c.Status(http.StatusBadGateway)
		fmt.Println(err)
		return
	}
	defer resp.Body.Close() //nolint:errcheck

	// If aux returned something other than 200, forward that status.
	if resp.StatusCode != http.StatusOK {
		c.Status(resp.StatusCode)
		return
	}

	var aux auxResponse
	if err := json.NewDecoder(resp.Body).Decode(&aux); err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, response{
		Version:    cfg.VERSION,
		AuxVersion: aux.Version,
		Data:       aux.Data,
	})
}

func livenessHandler(c *gin.Context) {
	c.Status(http.StatusOK)
}

func main() {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		log.Fatal(err)
	}

	r := gin.Default()

	r.GET("/buckets", func(c *gin.Context) {
		proxyHandler(c, "/buckets", cfg)
	})
	r.GET("/parameters", func(c *gin.Context) {
		proxyHandler(c, "/parameters", cfg)
	})
	r.GET("/parameters/:name", func(c *gin.Context) {
		name := c.Param("name")
		proxyHandler(c, "/parameters/"+name, cfg)
	})

	// Health entpoint
	r.GET("/livez", livenessHandler)

	addr := ":8080"
	log.Printf("Service listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("router error: %v", err)
	}
}
