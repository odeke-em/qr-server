package models

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

type Lease struct {
	Duration     int64
	AuthorizerId bson.ObjectId
	CreationDate time.Time
	PrivateKey   string
	PublicKey    string
}

func (l *Lease) Expired() bool {
	return l.CreationDate.Add(time.Duration(l.Duration)).Before(time.Now())
}
