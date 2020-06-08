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

package db

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/open-falcon/falcon-plus/modules/hbs/g"
)

var DB *sql.DB

var DBRO *sql.DB

func Init() {
	var err error
	DB, err = sql.Open("mysql", g.Config().Database)
	//DBRO ,errro = sql.Open("mysql", g.Config().DatabaseRo)
	if err != nil {
		g.Logger.Fatalf("open master db fail:", err)
	}

	DB.SetMaxOpenConns(g.Config().MaxConns)
	DB.SetMaxIdleConns(g.Config().MaxIdle)

	err = DB.Ping()
	if err != nil {
		g.Logger.Fatalf("ping master db fail:", err)
	}
	SlaveInit()

}

func SlaveInit() {
	var err error
	DBRO, err = sql.Open("mysql", g.Config().DatabaseRo)
	//DBRO ,errro = sql.Open("mysql", g.Config().DatabaseRo)
	if err != nil {
		g.Logger.Fatalf("open master db fail:", err)
	}

	DBRO.SetMaxOpenConns(g.Config().MaxConns)
	DBRO.SetMaxIdleConns(g.Config().MaxIdle)

	err = DBRO.Ping()

	if err != nil {
		g.Logger.Fatalf("ping master db fail:", err)
	}

}
