package main

import (
	"encoding/base64"
	"fmt"
	"github.com/go-redis/redis"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func main() {
	redisHost, redisAuth := initParams()
	dataCache := NewDataCache(redisHost, redisAuth)
	fs := http.FileServer(http.Dir("./"))
	http.Handle("/busuanzi.pure.mini.js", fs)
	http.Handle("/bsz.pure.mini.min.js", fs)
	http.Handle("/bsz.pure.mini.js", fs)
	http.Handle("/busuanzi", dataCache)
	http.Handle("/", dataCache)

	server := http.Server{
		Addr: "0.0.0.0:18080",
	}
	server.ListenAndServe()
}

func initParams() (string, string) {
	redisHost := os.Getenv("REDIS_HOST")
	redisAuth := os.Getenv("REDIS_AUTH")
	domain := os.Getenv("DOMAIN")
	jsCodeByte, err := ioutil.ReadFile("./busuanzi.js")
	jsCode := string(jsCodeByte)
	if err != nil {
		jsCode = "ERROR\n"
	} else {
		jsCode = strings.ReplaceAll(jsCode, "busuanzi.ibruce.info/busuanzi", domain) + "\n"
	}
	ioutil.WriteFile("./bsz.pure.mini.js", []byte(jsCode), 0666)
	ioutil.WriteFile("./bsz.pure.mini.min.js", []byte(jsCode), 0666)
	ioutil.WriteFile("./busuanzi.pure.mini.js", []byte(jsCode), 0666)
	return redisHost, redisAuth
}

type CacheData struct {
	http.Handler
	client *redis.Client
}

func NewDataCache(host string, auth string) *CacheData {
	c := CacheData{}
	c.client = redis.NewClient(&redis.Options{
		Addr:     host,
		Password: auth,
		DB:       0,
	})
	return &c
}

func (r CacheData) ServeHTTP(resW http.ResponseWriter, req *http.Request) {
	referer := req.Header.Get("Referer")
	if referer == "" {
		resW.Write([]byte("{}\n"))
		return
	}
	err := req.ParseForm()
	if err != nil {
		resW.Write([]byte("ERROR"))
	}
	up, err := url.Parse(referer)
	jsonpCallback := req.FormValue("jsonpCallback")

	remoteIp := ""
	//获取客户端IP
	if req.Header.Get("X-Real-IP") != "" {
		remoteIp = req.Header.Get("X-Real-IP")
	}
	if remoteIp == "" && req.Header.Get("X-Forwarded-For") != "" {
		remoteIp = req.Header.Get("X-Forwarded-For")
	}
	if remoteIp == "" {
		remoteIp = req.RemoteAddr
	}

	sitePV, siteUV, pagePV := r.getPVUV(up.Host, up.Path, remoteIp)
	rsStr := fmt.Sprintf("try{%s({\"site_uv\":%d,\"page_pv\":%d,\"version\":2.3,\"site_pv\":%d});}catch(e){}\n",
		jsonpCallback, siteUV, pagePV, sitePV)
	resW.Write([]byte(rsStr))
}

func (r CacheData) getPVUV(host string, url string, ip string) (int64, int64, int64) {
	hostBase64 := base64.StdEncoding.EncodeToString([]byte(host))
	pvKey := hostBase64 + ":PV"
	uvKey := hostBase64 + ":UV"
	urlKey := base64.StdEncoding.EncodeToString([]byte(url))

	if strings.Contains(ip, ":") {
		ip = strings.Split(ip, ":")[0]
	}

	var reSitePV int64 = 0
	var rePagePV int64 = 0
	var reSiteUV int64 = 0

	reSitePV, err := r.client.HIncrBy(pvKey, "/__SITE_ALL_COUNT__", 1).Result()
	if err != nil {
		fmt.Println(err)
	}
	rePagePV, err = r.client.HIncrBy(pvKey, urlKey, 1).Result()
	if err != nil {
		fmt.Println(err)
	}
	_, err = r.client.PFAdd(uvKey, ip).Result()
	if err != nil {
		fmt.Println(err)
	}
	reSiteUV, err = r.client.PFCount(uvKey).Result()
	if err != nil {
		fmt.Println(err)
	}
	return reSitePV, reSiteUV, rePagePV
}
