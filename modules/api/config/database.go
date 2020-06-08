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

package config

import (
	"database/sql"
	"fmt"

	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/spf13/viper"
)

type DBPool struct {
	Falcon    *gorm.DB
	Graph     *gorm.DB
	Uic       *gorm.DB
	Dashboard *gorm.DB
	Alarm     *gorm.DB
}

var (
	dbp DBPool
	//dbp_ro DBPool
)

func Con() DBPool {
	return dbp
}

//func ConRo() DBPool {
//	return dbp_ro
//}

func SetLogLevel(loggerlevel bool) {
	dbp.Uic.LogMode(loggerlevel)
	dbp.Graph.LogMode(loggerlevel)
	dbp.Falcon.LogMode(loggerlevel)
	dbp.Dashboard.LogMode(loggerlevel)
	dbp.Alarm.LogMode(loggerlevel)

	//dbp_ro.Uic.LogMode(loggerlevel)
	//dbp_ro.Graph.LogMode(loggerlevel)
	//dbp_ro.Falcon.LogMode(loggerlevel)
	//dbp_ro.Dashboard.LogMode(loggerlevel)
	//dbp_ro.Alarm.LogMode(loggerlevel)
}

func InitDB(loggerlevel bool) (err error) {

	//init redis
	log.Println("redis_config ", viper.GetString("redis.addr"))
	log.Println("redis_config ", viper.GetString("redis.maxIdle"))
	InitRedisConnPool(viper.GetString("redis.addr"), viper.GetInt("redis.maxIdle"))

	//portal master
	var p *sql.DB
	portal, err := gorm.Open("mysql", viper.GetString("db.falcon_portal"))
	portal.Dialect().SetDB(p)
	portal.LogMode(loggerlevel)
	if err != nil {
		return fmt.Errorf("connect to falcon_portal: %s", err.Error())
	}
	portal.SingularTable(true)
	dbp.Falcon = portal

	////portal slave
	//var pro *sql.DB
	//portalro, err := gorm.Open("mysql", viper.GetString("db.falcon_portal"))
	//portalro.Dialect().SetDB(pro)
	//portalro.LogMode(loggerlevel)
	//if err != nil {
	//	return fmt.Errorf("connect to db.falcon_portal: %s", err.Error())
	//}
	//portalro.SingularTable(true)
	//dbp_ro.Falcon = portalro

	//graph master
	var g *sql.DB
	graphd, err := gorm.Open("mysql", viper.GetString("db.graph"))
	graphd.Dialect().SetDB(g)
	graphd.LogMode(loggerlevel)
	if err != nil {
		return fmt.Errorf("connect to graph: %s", err.Error())
	}
	graphd.SingularTable(true)
	dbp.Graph = graphd

	//graph slave
	//var gro *sql.DB
	//graphdro, err := gorm.Open("mysql", viper.GetString("db.graph"))
	//graphdro.Dialect().SetDB(gro)
	//graphdro.LogMode(loggerlevel)
	//if err != nil {
	//	return fmt.Errorf("connect to db.graph: %s", err.Error())
	//}
	//graphdro.SingularTable(true)
	//dbp_ro.Graph = graphdro

	//uic master
	var u *sql.DB
	uicd, err := gorm.Open("mysql", viper.GetString("db.uic"))
	uicd.Dialect().SetDB(u)
	uicd.LogMode(loggerlevel)
	if err != nil {
		return fmt.Errorf("connect to uic: %s", err.Error())
	}
	uicd.SingularTable(true)
	dbp.Uic = uicd

	//uic slave
	//var uro *sql.DB
	//uicdro, err := gorm.Open("mysql", viper.GetString("db.uic"))
	//uicdro.Dialect().SetDB(uro)
	//uicdro.LogMode(loggerlevel)
	//if err != nil {
	//	return fmt.Errorf("connect to db.uic: %s", err.Error())
	//}
	//uicdro.SingularTable(true)
	//dbp_ro.Uic = uicdro

	//dashboard master
	var d *sql.DB
	dashd, err := gorm.Open("mysql", viper.GetString("db.dashboard"))
	dashd.Dialect().SetDB(d)
	dashd.LogMode(loggerlevel)
	if err != nil {
		return fmt.Errorf("connect to dashboard: %s", err.Error())
	}
	dashd.SingularTable(true)
	dbp.Dashboard = dashd

	//dashboard slave
	//var dro *sql.DB
	//dashdro, err := gorm.Open("mysql", viper.GetString("db.dashboard"))
	//dashdro.Dialect().SetDB(dro)
	//dashdro.LogMode(loggerlevel)
	//if err != nil {
	//	return fmt.Errorf("connect to db.dashboard: %s", err.Error())
	//}
	//dashdro.SingularTable(true)
	//dbp_ro.Dashboard = dashdro

	//alarm master
	var alm *sql.DB
	almd, err := gorm.Open("mysql", viper.GetString("db.alarms"))
	almd.Dialect().SetDB(alm)
	almd.LogMode(loggerlevel)
	if err != nil {
		return fmt.Errorf("connect to alarms: %s", err.Error())
	}
	almd.SingularTable(true)
	dbp.Alarm = almd

	//alarm slave
	//var almro *sql.DB
	//almdro, err := gorm.Open("mysql", viper.GetString("db.alarms"))
	//almdro.Dialect().SetDB(almro)
	//almdro.LogMode(loggerlevel)
	//if err != nil {
	//	return fmt.Errorf("connect to db.alarms: %s", err.Error())
	//}
	//almdro.SingularTable(true)
	//dbp_ro.Alarm = almdro

	//init redis
	//InitRedisConnPool()

	SetLogLevel(loggerlevel)
	return
}

func CloseDB() (err error) {
	err = dbp.Falcon.Close()
	if err != nil {
		return
	}
	err = dbp.Graph.Close()
	if err != nil {
		return
	}
	err = dbp.Uic.Close()
	if err != nil {
		return
	}
	err = dbp.Dashboard.Close()
	if err != nil {
		return
	}
	err = dbp.Alarm.Close()
	if err != nil {
		return
	}
	return
}
