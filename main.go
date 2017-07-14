package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"sync"

	"strings"

	"github.com/riclava/h2t/handler"
	"gopkg.in/gin-gonic/gin.v1"
)

type apiHandler struct {
	sync.RWMutex
	acl *handler.ACLHandler
}

func loadACL() *handler.ACLHandler {
	config, e := ioutil.ReadFile(*cfg)
	if e != nil {
		log.Fatal("Unable to load services config file", e)
		return nil
	}
	if *debug {
		log.Println("loading config file [" + *cfg + "]")
	}

	acl := handler.ACLHandler{}

	json.Unmarshal(config, &acl)
	if *debug {
		log.Println("Loaded rules ", acl)
	}
	return &acl
}

func (api *apiHandler) items(c *gin.Context) {
	api.RLock()
	defer api.RUnlock()
	c.JSON(200, api.acl.Copy())
}

func (api *apiHandler) put(c *gin.Context) {
	api.Lock()
	defer api.Unlock()

	req := handler.Service{}
	if err := c.Bind(&req); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	if strings.Compare(req.Name, "") == 0 {
		c.AbortWithStatus(http.StatusNotAcceptable)
		return
	}
	err := api.acl.Put(req.Name, req.Date, req.Description)
	if err != nil {
		c.AbortWithError(http.StatusNotAcceptable, err)
	} else {
		c.AbortWithStatus(http.StatusNoContent)
	}
}

func (api *apiHandler) delete(c *gin.Context) {
	api.Lock()
	defer api.Unlock()
	err := api.acl.Delete(c.Param("service"))
	if err != nil {
		c.AbortWithStatus(http.StatusNotFound)
	} else {
		c.AbortWithStatus(http.StatusNoContent)
	}

}

func (api *apiHandler) deleteAll(c *gin.Context) {
	api.Lock()
	defer api.Unlock()
	api.acl.DeleteAll()
	c.AbortWithStatus(http.StatusNoContent)
}

// flush memory config to disk
func (api *apiHandler) flush(c *gin.Context) {
	cfgBytes, err := json.Marshal(api.acl)
	if err != nil {
		log.Fatal("Unable to parse services to bytes", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	err = ioutil.WriteFile(*cfg, cfgBytes, 0644)
	if err != nil {
		log.Fatal("Unable to write services to config file", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.AbortWithStatus(http.StatusOK)
}

var bind = flag.String("bind", ":8081", "Binding address")
var withAPI = flag.Bool("with-api", true, "Allow service config API on HTTP")
var debug = flag.Bool("debug", false, "Is http2tcp running on debug mode")
var cfg = flag.String("config", "./conf/services.json", "Config filename")

type router struct {
	Proxy http.Handler
	Other http.Handler
}

func (rt *router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "CONNECT" {
		rt.Proxy.ServeHTTP(w, r)
	} else {
		rt.Other.ServeHTTP(w, r)
	}
}

func main() {
	flag.Parse()

	if len(flag.Args()) == 0 {
		if *debug {
			log.Println("Using default config")
		}
	}

	aclHandler := loadACL()

	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()

	if *withAPI {
		handler := &apiHandler{acl: aclHandler}
		r.GET("/api", handler.items)
		r.DELETE("/api", handler.deleteAll)
		r.POST("/api", handler.put)
		r.DELETE("/api/:service", handler.delete)
		r.PUT("/api", handler.flush)
	}
	log.Fatal(http.ListenAndServe(*bind, &router{Proxy: aclHandler, Other: r}))
}
