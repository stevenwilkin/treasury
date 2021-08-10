package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
)

const (
	socketPath = "/tmp/treasuryd.sock"
)

var client = http.Client{
	Transport: &http.Transport{
		DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
			return net.Dial("unix", socketPath)
		},
	},
}

func get(path string, result interface{}) {
	resp, err := client.Get(fmt.Sprintf("http://unix%s", path))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	json.Unmarshal(body, result)
	if err != nil {
		panic(err)
	}
}

func main() {
	rootCmd.Execute()
}
