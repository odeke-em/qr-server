package models

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

const (
	DbName          = "qr-server"
	KeySetModelName = "keyset"
)

type User struct {
	lastEditTime    time.Time
	Email           string
	MainPublicKey   string
	MainPrivateKey  string
	ExpiryDateEpoch int64
	id              bson.ObjectId `bson:"_id,omitempty"`
}

type KeySet struct {
	LastEditTime time.Time `json:"lastedit_time"`
	DateCreated  time.Time `json:"date_created"`
	PublicKey    string    `json:"public_key"`
	PrivateKey   string    `json:"private_key"`
	ownerId      *bson.ObjectId
	id           bson.ObjectId `bson:"_id,omitempty"`
}

func currentTimeUTC() time.Time {
	return time.Now().UTC()
}

func (ks *KeySet) PreSave() {
	if !ks.id.Valid() {
		ks.id = bson.NewObjectId()
	}
	ks.LastEditTime = currentTimeUTC()
}

func (ks *KeySet) Init() {
	if !ks.id.Valid() {
		ks.id = bson.NewObjectId()
	}
	ks.DateCreated = currentTimeUTC()
}

func (u *User) Valid() bool {
	// TODO: check email properly
	return time.Now().Unix() < u.ExpiryDateEpoch
}
