package main

import (
	"fmt"
	"net/http"
	"net/url"

	// drive "github.com/odeke-em/drive/src"
	"github.com/odeke-em/meddler"
)

func googleDriveDomainRestrictedQRCode(pl meddler.Payload, res http.ResponseWriter, req *http.Request) {
	uri := pl.URI

	parsedURL, err := url.Parse(uri)
	if err != nil {
		http.Error(res, fmt.Sprintf("parseURL %q got %v", uri, err), 500)
		return
	}

	// Uncomment once godep starts working again
	// resourceURLStr := drive.DriveResourceEntryURL -- godep is tripping with drive, as of `Sat Oct 31 04:14:35 MDT 2015`
	resourceURLStr := "https://drive.google.com"
	driveParsedURL, dErr := url.Parse(resourceURLStr)

	if dErr != nil {
		panic(fmt.Errorf("driveResourceEntryURL %q got %v", resourceURLStr, dErr))
	}

	if parsedURL.Host != driveParsedURL.Host {
		errMsg := fmt.Sprintf("expecting only urls to host %q not %q", driveParsedURL.Host, parsedURL.Host)
		http.Error(res, errMsg, 403)
		return
	}

	presentQRCode(pl, res, req)
}
