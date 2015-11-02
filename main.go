package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/binding"

	"github.com/odeke-em/extractor"
	uuid "github.com/odeke-em/go-uuid"
	"github.com/odeke-em/meddler"
	"github.com/odeke-em/qr-server/models"
	"github.com/odeke-em/rsc/qr"
)

const (
	ENV_DRIVE_SERVER_PUB_KEY  = "DRIVE_SERVER_PUB_KEY"
	ENV_DRIVE_SERVER_PRIV_KEY = "DRIVE_SERVER_PRIV_KEY"
	ENV_DRIVE_SERVER_PORT     = "DRIVE_SERVER_PORT"
	ENV_DRIVE_SERVER_HOST     = "DRIVE_SERVER_HOST"

	ENV_DBNAME           = "DBNAME"
	ENV_RESTRICT_DOMAINS = "RESTRICT_DOMAINS"

	DefaultMongoURI = "mongodb://localhost:27017"
	DefaultPort     = "4040"
)

var envKeyAlias = &extractor.EnvKey{
	PubKeyAlias:  ENV_DRIVE_SERVER_PUB_KEY,
	PrivKeyAlias: ENV_DRIVE_SERVER_PRIV_KEY,
}

type addressInfo struct {
	port, host string
}

func envGet(varname string, placeholders ...string) string {
	v := os.Getenv(varname)
	if v == "" {
		for _, placeholder := range placeholders {
			if placeholder != "" {
				v = placeholder
				break
			}
		}
	}

	return v
}

func mongoURI() string {
	uri := os.Getenv("MONGOHQ_URI")
	if uri == "" {
		uri = DefaultMongoURI
	}

	return uri
}

func addressInfoFromEnv() *addressInfo {
	return &addressInfo{
		port: envGet(ENV_DRIVE_SERVER_PORT, "3000"),
		host: envGet(ENV_DRIVE_SERVER_HOST, "localhost"),
	}
}

var envKeySet = extractor.KeySetFromEnv(envKeyAlias)
var envAddrInfo = addressInfoFromEnv()
var envDbName = envGet(ENV_DBNAME, models.DbName)

func (ai *addressInfo) ConnectionString() string {
	// TODO: ensure fields meet rubric
	return fmt.Sprintf("%s:%s", ai.host, ai.port)
}

func main() {
	m := martini.Classic()

	if envGet(ENV_RESTRICT_DOMAINS) == "" {
		m.Get("/qr", binding.Bind(meddler.Payload{}), presentQRCode)
		m.Post("/qr", binding.Bind(meddler.Payload{}), presentQRCode)
	}

	m.Get("/drive/qr", binding.Bind(meddler.Payload{}), googleDriveDomainRestrictedQRCode)
	m.Post("/drive/qr", binding.Bind(meddler.Payload{}), googleDriveDomainRestrictedQRCode)

	// m.Post("/gen", GenerateKeySet)

	m.Run() // m.RunOnAddr(envAddrInfo.ConnectionString())
}

func sessionHandler(fn func(*mgo.Session) (interface{}, error)) (interface{}, error) {
	uri := mongoURI()

	session, sErr := mgo.Dial(uri)
	if sErr == nil {
		defer session.Close()
		session.SetSafe(&mgo.Safe{})
	}

	result, err := fn(session)

	return result, err
}

func lookUpKeySet(publicKey string) (*extractor.KeySet, error) {
	result, err := sessionHandler(func(session *mgo.Session) (interface{}, error) {
		collection := session.DB(envDbName).C(models.KeySetModelName)
		result := models.KeySet{}

		if qErr := collection.Find(bson.M{"publickey": publicKey}).One(&result); qErr != nil {
			return nil, qErr
		}

		ks := &extractor.KeySet{
			PublicKey:  result.PublicKey,
			PrivateKey: result.PrivateKey,
		}

		return ks, nil
	})

	ks, _ := result.(*extractor.KeySet)

	return ks, err
}

func GenerateKeySet(res http.ResponseWriter, req *http.Request) {
	ks, err := _generateKeySet()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v error: %v\n", time.Now().UTC(), err)
		fmt.Fprintf(res, "error encountered please try again!")
		return
	}

	marshalled, err := json.Marshal(ks)
	var result interface{} = string(marshalled)

	if err != nil {
		result = err
	}

	fmt.Fprintf(res, "%v\n", result)
}

func newUUID4Joined() string {
	return strings.Replace(uuid.NewRandom().String(), "-", "", -1)
}

func newKeySet() *models.KeySet {
	ks := &models.KeySet{
		PublicKey:  newUUID4Joined(),
		PrivateKey: newUUID4Joined(),
	}

	ks.Init()

	return ks
}

func _generateKeySet() (*models.KeySet, error) {
	result, err := sessionHandler(func(session *mgo.Session) (interface{}, error) {
		ks := newKeySet()

		collection := session.DB(envDbName).C(models.KeySetModelName)
		index := mgo.Index{
			Key:        []string{"privatekey", "publickey"},
			Unique:     true,
			Background: false, // TODO: check if blocking will affect speed
		}

		indexErr := collection.EnsureIndex(index)
		if indexErr != nil {
			return nil, indexErr
		}

		ks.PreSave()
		err := collection.Insert(ks)
		if err != nil {
			return nil, err
		}

		return ks, nil
	})

	mks, _ := result.(*models.KeySet)
	return mks, err
}

func presentQRCode(pl meddler.Payload, res http.ResponseWriter, req *http.Request) {
	foundKeySet, err := lookUpKeySet(pl.PublicKey)

	if err != nil {
		http.Error(res, fmt.Sprintf("encountered error: %v", err), 400)
		return
	}

	if foundKeySet == nil {
		panic("null keySet returned from publicKeyLookup")
	}

	rawTextForSigning := pl.RawTextForSigning()
	if !foundKeySet.Match([]byte(rawTextForSigning), []byte(pl.Signature)) {
		http.Error(res, "invalid signature", 403)
		return
	}

	curTimeUnix := time.Now().Unix()
	if pl.ExpiryTime < curTimeUnix {
		http.Error(res, fmt.Sprintf("request expired at %q, current time %q", pl.ExpiryTime, curTimeUnix), 403)
		return
	}

	uri := pl.URI
	code, err := qr.Encode(uri, qr.Q)
	if err != nil {
		fmt.Fprintf(res, "%s %v\n", uri, err)
		return
	}

	pngImage := code.PNG()
	fmt.Fprintf(res, "%s", pngImage)
}

func errorPrint(fmt_ string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "\033[31m")
	fmt.Fprintf(os.Stderr, fmt_, args...)
	fmt.Fprintf(os.Stderr, "\033[00m")
}
