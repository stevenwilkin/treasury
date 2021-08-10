package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
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

func post(path string, values url.Values) {
	resp, err := client.PostForm(fmt.Sprintf("http://unix%s", path), values)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Failed")
		os.Exit(1)
	}
}

func main() {
	rootCmd.Execute()
}
