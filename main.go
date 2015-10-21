package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/binding"

	"github.com/odeke-em/extractor"
	"github.com/odeke-em/meddler"
	"github.com/odeke-em/rsc/qr"
)

const (
	ENV_DRIVE_SERVER_PUB_KEY  = "DRIVE_SERVER_PUB_KEY"
	ENV_DRIVE_SERVER_PRIV_KEY = "DRIVE_SERVER_PRIV_KEY"
	ENV_DRIVE_SERVER_PORT     = "DRIVE_SERVER_PORT"
	ENV_DRIVE_SERVER_HOST     = "DRIVE_SERVER_HOST"
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

func addressInfoFromEnv() *addressInfo {
	return &addressInfo{
		port: envGet(ENV_DRIVE_SERVER_PORT, "3000"),
		host: envGet(ENV_DRIVE_SERVER_HOST, "localhost"),
	}
}

var envKeySet = extractor.KeySetFromEnv(envKeyAlias)
var envAddrInfo = addressInfoFromEnv()

func (ai *addressInfo) ConnectionString() string {
	// TODO: ensure fields meet rubric
	return fmt.Sprintf("%s:%s", ai.host, ai.port)
}

func main() {
	if envKeySet.PublicKey == "" {
		errorPrint("publicKey not set. Please set %s in your env.\n", envKeyAlias.PubKeyAlias)
		return
	}

	if envKeySet.PrivateKey == "" {
		errorPrint("privateKey not set. Please set %s in your env.\n", envKeyAlias.PrivKeyAlias)
		return
	}

	m := martini.Classic()

	m.Get("/qr", binding.Bind(meddler.Payload{}), presentQRCode)
	m.Post("/qr", binding.Bind(meddler.Payload{}), presentQRCode)

	m.Run() // m.RunOnAddr(envAddrInfo.ConnectionString())
}

func presentQRCode(pl meddler.Payload, res http.ResponseWriter, req *http.Request) {
	if pl.PublicKey != envKeySet.PublicKey {
		http.Error(res, "invalid publickey", 405)
		return
	}

	rawTextForSigning := pl.RawTextForSigning()
	if !envKeySet.Match([]byte(rawTextForSigning), []byte(pl.Signature)) {
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
