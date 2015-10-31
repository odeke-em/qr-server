package main

import (
    "fmt"
    "io/ioutil"
    "net/http"
    "os"
)

func main() {
    uri := "http://localhost:3000/gen"
    if len(os.Args) > 1 { 
        uri = os.Args[1]
    }
    response, err := http.Post(uri, "", nil)
    if err != nil {
        panic(err)
    }

    if response == nil {
        panic("nil response encountered")
    }

    body := response.Body

	defer body.Close()
	data, rErr := ioutil.ReadAll(body)
    if rErr != nil {
        panic(rErr)
    }

    fmt.Printf("%s", data)
}
