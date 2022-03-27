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

var client, redisopenerr = rueidis.NewClient(rueidis.ClientOption{
	InitAddress: []string{"127.0.0.1:6479"},
})
var ctx = context.Background()

func checkurl(url string) bool {
	c := colly.NewCollector()
	var getStatusCode int
	c.OnResponse(func(r *colly.Response) {
		getStatusCode = r.StatusCode
	})
	c.OnError(func(r *colly.Response, err error) {
		getStatusCode = r.StatusCode
	})
	c.Visit(url)
	if getStatusCode == 200 {
		return true
	} else {
		return false
	}
}

const timeLayout = time.RFC3339

var newTaipeiZone, _ = time.LoadLocation("Asia/Taipei")

func Api(c *gin.Context) {

	data := map[string]string{}
	err := c.Bind(&data)
	if err != nil {
		fmt.Printf("data %v\n", data)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "請輸入正確資料",
			"msg":   err.Error(),
		})
		return
	}

	var dataurl string = data["url"]
	var expireAt string = data["expireAt"]
	fmt.Println(dataurl)
	fmt.Println(expireAt)
	if !checkurl(dataurl) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "錯誤網址",
		})
		return
	}

	timenow := time.Now().UTC().In(newTaipeiZone)
	convert_expireAt, err := time.Parse(timeLayout, expireAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "無法分析日期",
		})
		return
	}
	timeresult := convert_expireAt.Sub(timenow)

	continuedtime_secounds := int64(timeresult.Seconds())

	client.Do(ctx, client.B().Incr().Key("id").Build())
	getid, _ := client.Do(ctx, client.B().Get().Key("id").Build()).ToString()
	client.Do(ctx, client.B().Set().Key(getid).Value(dataurl).Nx().Build()).Error()
	client.Do(ctx, client.B().Expire().Key(getid).Seconds(continuedtime_secounds).Build())
	urlresult := "http://localhost:/" + getid

	c.JSON(http.StatusBadRequest, gin.H{
		"id":       getid,
		"shortUrl": urlresult,
	})

}

func Redirectapi(c *gin.Context) {

	if _, err := strconv.ParseInt(c.Param("urlid"), 10, 0); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"result": "輸入錯誤參數",
		})
		return
	}

	geturl, err := client.Do(ctx, client.B().Get().Key(c.Param("urlid")).Build()).ToString()

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"result": "資源已過期",
		})
		return
	} else {
		c.Redirect(http.StatusMovedPermanently, geturl)
	}
}

func main() {
	defer client.Close()
	if redisopenerr != nil {
		fmt.Println("redis 連接失敗")
		return
	}
	client.Do(ctx, client.B().Set().Key("id").Value("0").Nx().Build()).Error()
	router := gin.Default()
	router.POST("/api/v1/urls", Api)
	router.GET("/:urlid", Redirectapi)
	router.Run("")
}
