package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/tidwall/gjson"
)

const searchUrl = "https://www.notion.so/api/v3/search"
const bodyTemplate = `{"sort": "Relevance", "spaceId": "%s", "source": "quick_find", "limit": 9, "filters": {"isDeletedOnly": false, "ancestors": [], "isNavigableOnly": true, "excludeTemplates": false, "lastEditedTime": [], "editedBy": [], "createdBy": [], "createdTime": [], "requireEditPermissions": false}, "query": "%s", "type": "BlocksInSpace"}`

func ReadEnv() EnvSetting {
	cookie := os.Getenv("cookie")

	namespace := os.Getenv("notionSpaceId")

	if cookie == "" || namespace == "" {
		panic("Some ENV variables are not set")
	}

	return EnvSetting{cookie, namespace}
}

func NotionSearch(env EnvSetting, query string) []byte {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	body := MakeBodyForQuery(env, query)
	req, _ := http.NewRequest("POST", searchUrl, body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", env.cookie)
	resp, _ := http.DefaultClient.Do(req)
	bytes, _ := ioutil.ReadAll(resp.Body)
	return bytes
}

func MakeBodyForQuery(env EnvSetting, query string) io.Reader {
	s := fmt.Sprintf(bodyTemplate, env.namespace, query)
	reader := strings.NewReader(s)
	return reader
}

type EnvSetting struct {
	cookie, namespace string
}

type ResponseItem struct {
	Id       string
	Title    string
	Subtitle string
	Icon     string
}

func NotionParseSearchResponse(resp []byte) []ResponseItem {
	result := make([]ResponseItem, 0)

	results := gjson.Get(string(resp), "results")
	records := gjson.Get(string(resp), "recordMap.block")

	for _, item := range results.Array() {
		resItem := ResponseItem{}
		resItem.Id = item.Get("id").String()
		resItem.Title = item.Get("highlight.text").String()
		resItem.Subtitle = item.Get("highlight.pathText").String()

		// clean up
		resItem.Subtitle = strings.ReplaceAll(resItem.Subtitle, "<gzkNfoUU>", "*")
		resItem.Subtitle = strings.ReplaceAll(resItem.Subtitle, "</gzkNfoUU>", "*")

		record := records.Get(resItem.Id)
		if record.Get("value.properties").Exists() {
			resItem.Title = record.Get("value.properties.title.0.0").String()
		}
		resItem.Icon = record.Get("value.format.page_icon").String()

		result = append(result, resItem)
	}
	return result
}
