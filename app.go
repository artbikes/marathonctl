package main

// All actions under the app command

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"strconv"
)

type AppList struct {
	client *Client
	format Formatter
}

func (a AppList) Apply(args []string) {
	path := "/v2/apps"
	request := a.client.GET(path)
	response, e := a.client.Do(request)
	Check(e == nil, "failed to get response", e)
	defer response.Body.Close()

	f := a.format.Format(response.Body, a.Humanize)
	fmt.Println(f)
}

func (a AppList) Humanize(body io.Reader) string {
	dec := json.NewDecoder(body)
	var applications Applications
	e := dec.Decode(&applications)
	Check(e == nil, "failed to unmarshal response", e)
	title := "APP VERSION USER\n"
	var b bytes.Buffer
	for _, app := range applications.Apps {
		b.WriteString(app.ID)
		b.WriteString(" ")
		b.WriteString(app.Version)
		b.WriteString(" ")
		b.WriteString(app.User)
		b.WriteString("\n")
	}
	text := title + b.String()
	return Columnize(text)
}

type AppVersions struct {
	client *Client
}

func (a AppVersions) Apply(args []string) {
	Check(len(args) > 0, "must supply id")
	id := url.QueryEscape(args[0])
	path := "/v2/apps/" + id + "/versions"
	request := a.client.GET(path)
	response, e := a.client.Do(request)
	Check(e == nil, "failed to list verions", e)
	dec := json.NewDecoder(response.Body)
	var versions Versions
	e = dec.Decode(&versions)
	Check(e == nil, "failed to unmarshal response", e)
	for _, version := range versions.Versions {
		fmt.Println(version)
	}
}

type AppShow struct {
	client *Client
}

func (a AppShow) Apply(args []string) {
	Check(len(args) == 2, "must provide id and version")
	id := url.QueryEscape(args[0])
	version := url.QueryEscape(args[1])
	path := "/v2/apps/" + id + "/versions/" + version
	request := a.client.GET(path)
	response, e := a.client.Do(request)
	Check(e == nil, "failed to show app", e)
	dec := json.NewDecoder(response.Body)
	var application Application
	e = dec.Decode(&application)
	Check(e == nil, "failed to unmarshal response", e)
	title := "INSTANCES MEM CMD\n"
	var b bytes.Buffer
	b.WriteString(strconv.Itoa(application.Instances))
	b.WriteString(" ")
	mem := fmt.Sprintf("%.2f", application.Mem)
	b.WriteString(mem)
	b.WriteString(" ")
	b.WriteString(application.Cmd)

	text := title + b.String()
	fmt.Println(Columnize(text))
}

type AppCreate struct {
	client *Client
}

func (a AppCreate) Apply(args []string) {
	Check(len(args) == 1, "must specifiy 1 jsonfile")

	f, e := os.Open(args[0])
	Check(e == nil, "failed to open jsonfile", e)
	defer f.Close()

	request := a.client.POST("/v2/apps", f)
	response, e := a.client.Do(request)
	Check(e == nil, "failed to get response", e)
	defer response.Body.Close()
	Check(response.StatusCode != 409, "app already exists")

	dec := json.NewDecoder(response.Body)
	var application Application
	e = dec.Decode(&application)
	Check(e == nil, "failed to decode response", e)
	fmt.Println(application.ID, application.Version)
}

type AppUpdate struct {
	client *Client
}

func (a AppUpdate) Apply(args []string) {
	Check(len(args) == 2, "must specify id and jsonfile")
	id := url.QueryEscape(args[0])
	f, e := os.Open(args[1])
	Check(e == nil, "failed to open jsonfile", e)
	defer f.Close()

	request := a.client.PUT("/v2/apps/"+id+"?force=true", f)
	response, e := a.client.Do(request)
	Check(e == nil, "failed to get response", e)
	defer response.Body.Close()

	sc := response.StatusCode
	Check(sc == 200, "bad status code", sc)

	dec := json.NewDecoder(response.Body)
	var update Update
	e = dec.Decode(&update)
	Check(e == nil, "failed to decode response", e)
	title := "DEPLOYID VERSION\n"
	text := title + update.DeploymentID + " " + update.Version
	fmt.Println(Columnize(text))
}

type AppRestart struct {
	client *Client
}

func (a AppRestart) Apply(args []string) {
	Check(len(args) == 1, "specify 1 app id to restart")
	id := url.QueryEscape(args[0])
	request := a.client.POST("/v2/apps/"+id+"/restart?force=true", nil)
	response, e := a.client.Do(request)
	Check(e == nil, "failed to get response", e)
	defer response.Body.Close()
	dec := json.NewDecoder(response.Body)
	var update Update
	e = dec.Decode(&update)
	Check(e == nil, "failed to decode response", e)
	title := "DEPLOYID VERSION\n"
	text := title + update.DeploymentID + " " + update.Version
	fmt.Println(Columnize(text))
}

type AppDestroy struct {
	client *Client
}

func (a AppDestroy) Apply(args []string) {
	Check(len(args) == 1, "must specify id")
	path := "/v2/apps/" + url.QueryEscape(args[0])
	request := a.client.DELETE(path)
	response, e := a.client.Do(request)
	Check(e == nil, "destroy app failed", e)
	c := response.StatusCode
	// documentation says this is 204, wtf
	Check(c == 200, "destroy app bad status", c)
	fmt.Println("destroyed", args[0])
}
