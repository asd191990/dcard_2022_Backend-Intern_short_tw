package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gavv/httpexpect/v2"
)

var urls = map[string]bool{
	"https://google.com":                             true,
	"https://ithelp.ithome.com.tw/articles/10212120": true,
	"wewcokl<ewqwee/":                                false,
	"asasas2324237wew.cwee":                          false,
}

func TestCheckUrlRegular(t *testing.T) {

	for UrlKey, _ := range urls {
		if CheckUrlRegular(UrlKey) {
			t.Log("success")
		} else {
			t.Error("fail")
		}
	}

}

func TestCheckurl(t *testing.T) {

	for UrlKey, ValueResult := range urls {
		if CheckUrl(UrlKey) == ValueResult {
			t.Log("success")
		} else {
			t.Error("fail")
		}
	}
}

func TestAddUrlData(t *testing.T) {
	url := "https://www.dcard.tw/f/talk/p/238474170"
	var continuedtime_secounds int64 = 600
	if getid, urlresult := AddUrlData(url, continuedtime_secounds); getid != "" && urlresult != "" {
		t.Log("success: ")
		t.Log("getid" + getid)
		t.Log("urlresult: " + urlresult)
	} else {
		t.Error("fail")
	}
}

func TestUrlCreateApi(t *testing.T) {
	Router := CreateRouter()
	server := httptest.NewServer(Router)
	defer server.Close()
	e := httpexpect.New(t, server.URL)

	orange := map[string]string{
		"url":      "https://www.dcard.tw/f/talk/p/238474170",
		"expireAt": "2022-03-31T03:26:29Z",
	}

	contentType := "application/json;charset=utf-8"

	e.POST("/api/v1/urls").
		WithHeader("ContentType", contentType).
		WithJSON(orange).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		ContainsKey("shortUrl")

	orange = map[string]string{
		"url":      "asojojklzx0",
		"expireAt": "2022-03-31T03:26:29Z",
	}

	e.POST("/api/v1/urls").
		WithHeader("ContentType", contentType).
		WithJSON(orange).
		Expect().
		Status(http.StatusBadRequest)

	// e.GET("/fruits").
	// 	Expect().
	// 	Status(http.StatusOK).JSON().Array().Empty()
}
