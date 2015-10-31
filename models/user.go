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
	lastEditTimeTrack
	Email           string
	MainPublicKey   string
	MainPrivateKey  string
	ExpiryDateEpoch int64
	id              bson.ObjectId `bson:"_id,omitempty"`
}

type KeySet struct {
	lastEditTimeTrack
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
	ownerId    *bson.ObjectId
	id         bson.ObjectId `bson:"_id,omitempty"`
}

func (ks *KeySet) SetupBeforeUpdate() {
	if !ks.id.Valid() {
		ks.id = bson.NewObjectId()
	}
	ks.lastEditTime = time.Now().Unix()
}

type lastEditTimeTrack struct {
	lastEditTime int64
}

func (u *User) Valid() bool {
	// TODO: check email properly
	return time.Now().Unix() < u.ExpiryDateEpoch
}
