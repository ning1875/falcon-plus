package g

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
)

func IpToEndpoint(ip string) (end string) {
	res := strings.Split(ip, ".")
	BlanStr := res[1]
	ClanStr := res[2]
	DlanStr := res[3]
	Clan, _ := strconv.ParseInt(ClanStr, 10, 64)
	Dlan, _ := strconv.ParseInt(DlanStr, 10, 64)

	if Clan < 100 {
		tmp := fmt.Sprintf("0%d", Clan)
		ClanStr = tmp
	}
	if Dlan < 100 {
		tmp := fmt.Sprintf("0%d", Dlan)
		DlanStr = tmp
	}
	return "n" + BlanStr + "-" + ClanStr + "-" + DlanStr
}

func GetAmsTagIps(tag string) []string {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
		}
	}()
	type Amstag struct {
		Response map[string]interface{} `json:"response"`
		Ret      int                    `json:"ret"`
	}

	v := url.Values{}
	v.Set("app_id", "135916235682081")
	v.Set("sign", "ff9406de0e8008f3308d4834e06583cb")
	v.Set("tag", tag)
	sendBody := bytes.NewReader([]byte(v.Encode()))

	send_url := "http://api-ops.xxx.com/ams/cluster/host_list"
	c := &http.Client{
		Timeout: 5 * time.Second,
	}
	req, _ := http.NewRequest("POST", send_url, sendBody)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	res, err := c.Do(req)
	if err != nil {
		log.Println(err)
		return nil
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
		return nil
	}
	amstag := Amstag{}
	err = json.Unmarshal(body, &amstag)
	if amstag.Ret != 0 {
		log.Errorf("reqeust to ams tag api return error. %+v", amstag)
		return nil
	}
	hosts := amstag.Response[tag]

	tmp := hosts.(map[string]interface{})["host"]
	var all []string
	switch reflect.TypeOf(tmp).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(tmp)
		for i := 0; i < s.Len(); i++ {
			ip := s.Index(i).Interface().(string)
			end := IpToEndpoint(ip)
			all = append(all, end)
		}
	}

	if err != nil {
		log.Println(err)
		return nil
	}
	//fmt.Println(all)
	return all
}
