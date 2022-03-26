package main

import (
	"fmt"
	"net/http"
	"strconv"

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

type Url struct {
	Url      string `form:"url" json:"url"`
	ExpireAt string `form:"expireAt" json:"expireAt"`
}

func Api(c *gin.Context) {

	fmt.Printf("wew\n")
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
	fmt.Printf("%v\n", data)

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

	//key
	client.Do(ctx, client.B().Incr().Key("id").Build())
	getid, _ := client.Do(ctx, client.B().Get().Key("id").Build()).ToString()
	client.Do(ctx, client.B().Set().Key(getid).Value(dataurl).Nx().Build()).Error()
	client.Do(ctx, client.B().Expire().Key(getid).Seconds(600).Build())
	c.JSON(http.StatusBadRequest, gin.H{
		"id":       getid,
		"shortUrl": dataurl,
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
			"result": "找不到網址",
		})
		return
	} else {
		c.Redirect(http.StatusMovedPermanently, geturl)
	}
}

func CreateData() {
	client.Do(ctx, client.B().Set().Key("id").Value("0").Nx().Build()).Error()
	client.Do(ctx, client.B().Incr().Key("id").Build())
}

func main() {
	router := gin.Default()
	router.POST("/api/v1/urls", Api)
	router.GET("/:urlid", Redirectapi)
	defer client.Close()

	client.Do(ctx, client.B().Set().Key("id").Value("0").Nx().Build()).Error()

	if redisopenerr == nil {
		// fmt.Println("can open")
	}

	router.Run(":8000")
}
