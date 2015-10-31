package models

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

const (
	DbName          = "qr-demo-server"
	KeySetModelName = "keyset"
)

type User struct {
	Email           string
	MainPublicKey   string
	MainPrivateKey  string
	ExpiryDateEpoch int64
	Id              *bson.ObjectId
}

type KeySet struct {
	PublicKey  string
	PrivateKey string
	OwnerId    *bson.ObjectId
	Id         *bson.ObjectId
}

func (u *User) Valid() bool {
	return u.Email != "" && time.Now().Unix() < u.ExpiryDateEpoch
}
