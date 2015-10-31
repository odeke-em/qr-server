package main

import (
	"fmt"
	"net/http"
	"net/url"

	drive "github.com/odeke-em/drive/src"
	"github.com/odeke-em/meddler"
)

func googleDriveDomainRestrictedQRCode(pl meddler.Payload, res http.ResponseWriter, req *http.Request) {
	uri := pl.URI

	parsedURL, err := url.Parse(uri)
	if err != nil {
		http.Error(res, fmt.Sprintf("parseURL %q got %v", uri, err), 500)
		return
	}

	driveParsedURL, dErr := url.Parse(drive.DriveResourceEntryURL)
	if dErr != nil {
		panic(fmt.Errorf("driveResourceEntryURL %q got %v", drive.DriveResourceEntryURL, dErr))
	}

	if parsedURL.Host != driveParsedURL.Host {
		errMsg := fmt.Sprintf("expecting only urls to host %q not %q", driveParsedURL.Host, parsedURL.Host)
		http.Error(res, errMsg, 403)
		return
	}

	presentQRCode(pl, res, req)
}
