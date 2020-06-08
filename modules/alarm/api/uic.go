// Copyright 2017 Xiaomi, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package api

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/open-falcon/falcon-plus/modules/alarm/g"
	"github.com/open-falcon/falcon-plus/modules/api/app/model/uic"
	"github.com/patrickmn/go-cache"
	"github.com/toolkits/container/set"
	"github.com/toolkits/net/httplib"
)

type APIGetTeamOutput struct {
	uic.Team
	Users       []*uic.User `json:"users"`
	TeamCreator string      `json:"creator_name"`
}

type UsersCache struct {
	sync.RWMutex
	M map[string][]*uic.User
}

//start set byte wenxudong
type AmsOwner struct {
	Response Owners `json:"response"`
}

type Owners struct {
	Owners     []string `json:"owner"`
	OwnerGroup []string `json:"owner_group"`
}

//end set

var Users = &UsersCache{M: make(map[string][]*uic.User)}

// cache : key team name ,value  []*users
var UserCache = cache.New(g.CacheTTl, g.CacheTTl)

func (this *UsersCache) Get(team string) []*uic.User {
	this.RLock()
	defer this.RUnlock()
	val, exists := this.M[team]
	if !exists {
		return nil
	}

	return val
}

func (this *UsersCache) Set(team string, users []*uic.User) {
	this.Lock()
	defer this.Unlock()
	this.M[team] = users
}

func UsersOf(team string) []*uic.User {
	//users := CurlUic(team)

	if res, found := UserCache.Get(team); found {
		return res.([]*uic.User)
	} else {
		curlRes := HttpCurlUic(team)
		UserCache.Set(team, curlRes, cache.DefaultExpiration)
		return curlRes
	}

}

func GetUsers(teams string) map[string]*uic.User {
	userMap := make(map[string]*uic.User)
	arr := strings.Split(teams, ",")
	for _, team := range arr {
		if team == "" {
			continue
		}

		users := UsersOf(team)
		if users == nil {
			continue
		}

		for _, user := range users {
			userMap[user.Name] = user
		}
	}
	return userMap
}

// return phones, emails, IM
func ParseTeamsForLarkGroup(teams, LarkGroupId string) ([]string, []string) {
	if teams == "" && LarkGroupId == "" {
		return []string{}, []string{}
	}
	phoneSet := set.NewStringSet()
	mailSet := set.NewStringSet()
	if teams != "" {
		userMap := GetUsers(teams)

		for _, user := range userMap {
			phoneSet.Add(user.Name)
			mailSet.Add(user.Name + g.BYTEMAIL)
		}
	}
	if LarkGroupId != "" {
		phoneSet.Add(LarkGroupId)
		mailSet.Add(LarkGroupId + g.BYTEMAIL)
	}

	return phoneSet.ToSlice(), mailSet.ToSlice()
}

// return phones, emails, IM
func ParseTeams(teams string) ([]string, []string) {
	if teams == "" {
		return []string{}, []string{}
	}

	userMap := GetUsers(teams)
	phoneSet := set.NewStringSet()
	mailSet := set.NewStringSet()
	for _, user := range userMap {
		phoneSet.Add(user.Name)
		mailSet.Add(user.Name + g.BYTEMAIL)
	}
	return phoneSet.ToSlice(), mailSet.ToSlice()
}

func CurlUic(team string) []*uic.User {
	if team == "" {
		return []*uic.User{}
	}

	//调用ams 短信发送接口，所有redis 电话号码key用用户名代替
	// uic 不维护电话号码等信息，只维护人员信息
	// return  [{"Name":"wenxudong", "Email":"xxx", "Phone":"xxx"}, {"Name":"wangning", "Email":"xxx", "Phone":"xxx"}]
	//start change use ams
	//if strings.HasPrefix(team, "ams.") {
	//	tag := strings.Replace(team, "ams.", "", 1)
	//	uri := fmt.Sprintf("https://ams.xxx.com/api.php?tag=%s&method=owner.get&token=ba8dc5f2791adcb66e2f601163f3d427", tag)
	//	req := httplib.Get(uri).SetTimeout(5 * time.Second, 30 * time.Second)
	//	var info AmsOwner
	//	err := req.ToJson(&info)
	//	if err != nil {
	//		log.Printf("curl ams %s fail: %v", uri, err)
	//	}
	//	//log.Printf("amstag: %s", tag)
	//	//log.Printf("amsuri: %s", uri)
	//	//log.Printf("amsteam: %s", team)
	//	var team_users APIGetTeamOutput
	//	for _, uid := range info.Response.Owners {
	//		user := uic.User{
	//			Name:uid,
	//			Email:uid + "@xxx.com",
	//			Phone:uid,
	//		}
	//		team_users.Users = append(team_users.Users, &user)
	//
	//	}
	//	//log.Printf("team users: %s", team_users.Users)
	//	return team_users.Users
	//}
	//end change use ams
	uri := fmt.Sprintf("%s/api/v1/team/name/%s", g.Config().Api.PlusApi, team)
	req := httplib.Get(uri).SetTimeout(2*time.Second, 10*time.Second)
	token, _ := json.Marshal(map[string]string{
		"name": "falcon-alarm",
		"sig":  g.Config().Api.PlusApiToken,
	})
	req.Header("Apitoken", string(token))
	var team_users APIGetTeamOutput
	err := req.ToJson(&team_users)
	if err != nil {
		log.Errorf("curl falcon %s fail: %v", uri, err)
		return nil
	}
	return team_users.Users
}

func HttpCurlUic(team string) []*uic.User {
	if team == "" {
		return []*uic.User{}
	}
	retry_limit := 3
	r_s := 0
	for r_s < retry_limit {
		//uri := fmt.Sprintf("%s/api/v1/team/name/%s", g.Config().Api.PlusApi, team)
		uri := fmt.Sprintf("%s/api/v1/team/name/%s", g.Config().Api.MainApi, team)
		req := httplib.Get(uri).SetTimeout(2*time.Second, 30*time.Second)
		token, _ := json.Marshal(map[string]string{
			"name": "falcon-alarm",
			"sig":  g.Config().Api.PlusApiToken,
		})
		req.Header("Apitoken", string(token))
		var team_users APIGetTeamOutput
		err := req.ToJson(&team_users)
		if err != nil {
			log.Errorf("HttpCurlUic_curl_falcon_round_%v_uri:%s_failreason: %v", r_s, uri, err)
			r_s += 1
			time.Sleep(time.Second)
		} else {
			return team_users.Users
		}

	}
	log.Errorf("HttpCurlUic_finally_failed_for_team:%s", team)
	return nil
}
