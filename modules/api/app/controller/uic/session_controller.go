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

package uic

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	h "github.com/open-falcon/falcon-plus/modules/api/app/helper"
	"github.com/open-falcon/falcon-plus/modules/api/app/model/uic"
	"github.com/open-falcon/falcon-plus/modules/api/app/utils"
	"github.com/open-falcon/falcon-plus/modules/api/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type APILoginInput struct {
	Name      string `json:"name"  form:"name" binding:"required"`
	Password  string `json:"password"  form:"password" binding:"required"`
	AdminSalt string `json:"admin_salt" form:"admin_salt"`
}

func Login(c *gin.Context) {
	inputs := APILoginInput{}
	if err := c.Bind(&inputs); err != nil {
		h.JSONR(c, badstatus, "name or password is blank")
		return
	}
	name := inputs.Name
	//password := inputs.Password

	var user uic.User

	configAdminSalt := viper.GetString("admin_salt")
	if configAdminSalt == "" {
		configAdminSalt = "falcon_api_admin_salt_byte2019"
	}
	//db.Uic.Where(&user).Find(&user)
	s := db.Uic.Table("user").Where("name = ?", name).Scan(&user)
	if s.Error != nil && s.Error.Error() != "record not found" {
		h.JSONR(c, badstatus, s.Error)
		return

	} else if user.ID == 0 {
		h.JSONR(c, badstatus, "no such user")
		return
	}

	sig_redis_key := fmt.Sprintf("user_sig_%s", user.Name)
	sig_v, err := config.RedisGet(sig_redis_key)
	if err != nil {
		log.Error(err)
	}

	if sig_v != "" {
		//如果想获取admin的sig 需校验请求中的admin salt
		if user.IsAdmin() == true {
			log.Infof("configAdminSalt:%s,inputs.AdminSalt:%s", configAdminSalt, inputs.AdminSalt)
			if configAdminSalt != inputs.AdminSalt {
				h.JSONR(c, badstatus, "wrong admin salt")
				return
			}
		}
		log.Debug("get_old_sig_from_redis", sig_v)
		var session_old uic.Session

		err = json.Unmarshal([]byte(sig_v), &session_old)
		if err != nil {
			log.Error("json_load_session_error", err)
			return
		}
		resp_old := struct {
			Sig   string `json:"sig,omitempty"`
			Name  string `json:"name,omitempty"`
			Admin bool   `json:"admin"`
		}{session_old.Sig, user.Name, user.IsAdmin()}
		h.JSONR(c, resp_old)
		return
	}
	log.Debug("sig_v", sig_v)
	var session uic.Session
	session.Sig = utils.GenerateUUID()
	//session.Expired = 3600*24*30
	session.Expired = 0
	session.Uid = user.ID

	bs, err := json.Marshal(session)
	if err != nil {
		log.Error(err)
		return
	}
	s_err := config.RedisSet(sig_redis_key, string(bs), session.Expired)
	if s_err != nil {
		log.Error(s_err)
		h.JSONR(c, badstatus, s_err)
		return
	}

	resp := struct {
		Sig   string `json:"sig,omitempty"`
		Name  string `json:"name,omitempty"`
		Admin bool   `json:"admin"`
	}{session.Sig, user.Name, user.IsAdmin()}
	h.JSONR(c, resp)
	return
}

func Logout(c *gin.Context) {
	wsession, err := h.GetSession(c)
	if err != nil {
		h.JSONR(c, badstatus, err.Error())
		return
	}
	var session uic.Session
	var user uic.User
	db.Uic.Table("user").Where(uic.User{Name: wsession.Name}).Scan(&user)
	db.Uic.Table("session").Where("sig = ? AND uid = ?", wsession.Sig, user.ID).Scan(&session)

	if session.ID == 0 {
		h.JSONR(c, badstatus, "not found this kind of session in database.")
		return
	} else {
		r := db.Uic.Table("session").Delete(&session)
		if r.Error != nil {
			h.JSONR(c, badstatus, r.Error)
		}
		h.JSONR(c, "logout successful")
	}
	return
}

func AuthSession(c *gin.Context) {
	auth, err := h.SessionChecking(c)
	if err != nil || auth != true {
		h.JSONR(c, http.StatusUnauthorized, err)
		return
	}
	h.JSONR(c, "session is vaild!")
	return
}

func CreateRoot(c *gin.Context) {
	password := c.DefaultQuery("password", "")
	if password == "" {
		h.JSONR(c, badstatus, "password is empty, please check it")
		return
	}
	password = utils.HashIt(password)
	user := uic.User{
		Name:   "root",
		Passwd: password,
	}
	dt := db.Uic.Table("user").Save(&user)
	if dt.Error != nil {
		h.JSONR(c, badstatus, dt.Error)
		return
	}
	h.JSONR(c, "root created!")
	return
}
