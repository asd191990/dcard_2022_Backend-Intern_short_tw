package models

import "time"

type UrlShort struct {
	Id        int
	Url       string
	LimitTime time.Time
}
