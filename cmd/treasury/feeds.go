package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sort"

	"github.com/spf13/cobra"
)

type feedsResponse struct {
	Feeds map[string]bool
}

func (fr *feedsResponse) feeds() []string {
	feeds := make([]string, len(fr.Feeds))
	i := 0
	for feed, _ := range fr.Feeds {
		feeds[i] = feed
		i++
	}
	sort.Strings(feeds)

	return feeds
}

func (fr *feedsResponse) padding() int {
	var longest int

	for feed, _ := range fr.Feeds {
		if len(feed) > longest {
			longest = len(feed)
		}
	}

	return longest
}

var feedsCmd = &cobra.Command{
	Use:   "feeds",
	Short: "Retrieve data feeds",
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := client.Get("http://unix/feeds")
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}

		var fr feedsResponse

		json.Unmarshal(body, &fr)
		if err != nil {
			panic(err)
		}

		var status string
		padding := fr.padding()

		for _, feed := range fr.feeds() {
			if fr.Feeds[feed] {
				status = "Active"
			} else {
				status = "Inactive"
			}
			fmt.Printf("%-*s %s\n", padding, feed, status)
		}
	},
}
