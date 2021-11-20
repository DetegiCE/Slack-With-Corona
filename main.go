package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type CoronaInfo struct {
	Overview     Overview     `json:"overview"`
	CasesSummary CasesSummary `json:"casesSummary"`
}

type Overview struct {
	Current   [2]int `json:"current"`
	Recovered [2]int `json:"recovered"`
	Deceased  [2]int `json:"deceased"`
	Confirmed [2]int `json:"confirmed"`
}

type CasesSummary struct {
	Checking       int `json:"checking"`
	TotalCases     int `json:"totalCases"`
	YesterdayCases int `json:"yesterdayCases"`
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

	fmt.Println(coronaInfo)
	return coronaInfo
}

func Slack_Divider() map[string]interface{} {
	data := make(map[string]interface{})
	data["type"] = "divider"
	return data
}

func Slack_Mrkdwn(text string) map[string]interface{} {
	data := make(map[string]interface{})
	innerData := make(map[string]interface{})
	data["type"] = "section"
	innerData["type"] = "mrkdwn"
	innerData["text"] = text
	data["text"] = innerData
	return data
}

func Build_Current_Time() string {
	loc, _ := time.LoadLocation("Asia/Seoul")
	then := time.Now().In(loc)
	return fmt.Sprintf("%s %02d, %02d:%02d", then.Month(), then.Day(), then.Hour(), then.Minute())
}

func Build_Slack_Message(tod string, yes string) map[string]interface{} {
	block := make(map[string]interface{})
	var data [4]map[string]interface{}

	data[0] = Slack_Mrkdwn("*" + Build_Current_Time() + "* 기준 대한민국 코로나 현황입니다.")
	data[1] = Slack_Divider()
	data[2] = Slack_Mrkdwn(":hot_face:\n오늘 현재까지 확진자 수 *" + tod + "* 명\n어제 동시간 대비 *" + yes + "* 명")
	data[3] = Slack_Divider()

	block["blocks"] = data
	return block
}

func Post_Corona_Info(url string, coronaInfo CoronaInfo) {
	today := strconv.Itoa(coronaInfo.CasesSummary.TotalCases)
	yesterday := strconv.Itoa(coronaInfo.Overview.Current[1])
	data := Build_Slack_Message(today, yesterday)
	body, _ := json.Marshal(data)
	buff := bytes.NewBuffer(body)

	fmt.Println(buff)
	req, reqErr := http.NewRequest(http.MethodPost, url, buff)
	if reqErr != nil {
		log.Fatal(reqErr)
	}
	req.Header.Add("Content-Type", "application/json")

	client := http.Client{
		Timeout: time.Second * 2,
	}

	res, postErr := client.Do(req)
	if postErr != nil {
		log.Fatal(postErr)
	}
	if res.Body != nil {
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
