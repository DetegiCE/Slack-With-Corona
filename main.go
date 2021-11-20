package main

import (
	"bytes"
	"encoding/json"
	"github.com/joho/godotenv"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

type CoronaInfo struct {
	CasesSummary CasesSummary `json:"casesSummary"`
}

type CasesSummary struct {
	Checking       int `json:"checking"`
	TotalCases     int `json:"totalCases"`
	YesterdayCases int `json:"yesterdayCases"`
}

type SlackMessage struct {
	Text string `json:"text"`
}

func Dot_Env_Variable(key string) string {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
	}
	return os.Getenv(key)
}

func Get_Corona_Info(url string) CoronaInfo {
	client := http.Client{
		Timeout: time.Second * 2,
	}

	req, reqErr := http.NewRequest(http.MethodGet, url, nil)
	if reqErr != nil {
		log.Fatal(reqErr)
	}
	req.Header.Set("User-Agent", "XYZ/3.0")

	res, getErr := client.Do(req)
	if getErr != nil {
		log.Fatal(getErr)
	}
	if res.Body != nil {
		defer res.Body.Close()
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}

	coronaInfo := CoronaInfo{}
	jsonErr := json.Unmarshal(body, &coronaInfo)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	return coronaInfo
}

func Post_Corona_Info(url string, coronaInfo CoronaInfo) {
	data := make(map[string]interface{})
	data["text"] = coronaInfo.CasesSummary.TotalCases
	body, _ := json.Marshal(data)
	buff := bytes.NewBuffer(body)

	req, reqErr := http.NewRequest(http.MethodPost, url, buff)
	if reqErr != nil {
		log.Fatal(reqErr)
	}
	req.Header.Add("Content-Type", "application/json")

	client := http.Client{}
	res, postErr := client.Do(req)
	if postErr != nil {
		defer res.Body.Close()
	}
}

func Slack_With_Corona() {
	url := "https://apiv2.corona-live.com/stats.json"
	webhookUrl := Dot_Env_Variable("SLACK_WEBHOOK_URL")

	coronaInfo := Get_Corona_Info(url)
	Post_Corona_Info(webhookUrl, coronaInfo)
}

func main() {
	Slack_With_Corona()
}
