package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"context"

	"github.com/gin-gonic/gin"
	"github.com/gocolly/colly"
	"github.com/rueian/rueidis"
)

var Client, Redisopenerr = rueidis.NewClient(rueidis.ClientOption{
	InitAddress: []string{"127.0.0.1:6379"},
})
var Ctx = context.Background()

const TIMELAYOUT = time.RFC3339

var NewTaipeiZone, _ = time.LoadLocation("Asia/Taipei")

func Checkurl(url string) bool {
	c := colly.NewCollector()
	var getStatusCode int
	fmt.Println("url:", url)

	c.OnResponse(func(r *colly.Response) {
		fmt.Println("ok")
		getStatusCode = r.StatusCode
	})
	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("error")
		fmt.Println(err.Error())
		getStatusCode = r.StatusCode
	})
	c.Visit(url)
	fmt.Println("getStatusCode:", getStatusCode)
	if getStatusCode == 200 || getStatusCode == 403 {
		return true
	} else {
		return false
	}
}

func AddUrlData(url string, continuedtime_secounds int64) (string, string) {
	Client.Do(Ctx, Client.B().Incr().Key("id").Build())
	getid, _ := Client.Do(Ctx, Client.B().Get().Key("id").Build()).ToString()
	Client.Do(Ctx, Client.B().Set().Key(getid).Value(url).Nx().Build())
	Client.Do(Ctx, Client.B().Expire().Key(getid).Seconds(continuedtime_secounds).Build())
	urlresult := "http://localhost:/" + getid
	return getid, urlresult
}

// handler

func UrlCreateApi(c *gin.Context) {

	data := map[string]string{}

	if err := c.Bind(&data); err != nil {
		fmt.Printf("data %v\n", data)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "請輸入正確資料",
		})
		return
	}

	var dataurl string = data["url"]
	var expireAt string = data["expireAt"]

	if len(dataurl) > 768 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "網址長度過長",
		})
		return
	}

	if !Checkurl(dataurl) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "錯誤網址",
		})
		return
	}

	timenow := time.Now().UTC().In(NewTaipeiZone)
	convert_expireAt, err := time.Parse(TIMELAYOUT, expireAt)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "無法分析日期",
		})
		return
	}

	timeresult := convert_expireAt.Sub(timenow)
	continuedtime_secounds := int64(timeresult.Seconds())

	if continuedtime_secounds < 0 {
		continuedtime_secounds = 600
	}

	id, urlresult := AddUrlData(dataurl, continuedtime_secounds)

	c.JSON(http.StatusOK, gin.H{
		"id":       id,
		"shortUrl": urlresult,
	})

}

func RedirectApi(c *gin.Context) {

	if _, err := strconv.ParseInt(c.Param("urlid"), 10, 0); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"result": "輸入錯誤參數",
		})
		return
	}

	geturl, err := Client.Do(Ctx, Client.B().Get().Key(c.Param("urlid")).Build()).ToString()

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"result": "資源已過期",
		})
		return
	} else {
		c.Redirect(http.StatusMovedPermanently, geturl)
	}
}

func CreateRouter() *gin.Engine {
	Router := gin.Default()
	Router.POST("/api/v1/urls", UrlCreateApi)
	Router.GET("/:urlid", RedirectApi)
	return Router
}

func main() {
	if Redisopenerr != nil {
		fmt.Println("redis 連接失敗")
		fmt.Println(Redisopenerr.Error())
		return
	}
	defer Client.Close()
	Client.Do(Ctx, Client.B().Set().Key("id").Value("0").Nx().Build()).Error()
	Router := CreateRouter()
	Router.Run()
}
