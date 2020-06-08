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

package falcon_portal

////////////////////////////////////////////////////////////////////////////////////
// |id                    | int(10) unsigned | NO   | PRI | NULL    | auto_increment |
// | uic                  | varchar(255)     | NO   |     |         |                |
// | url                  | varchar(255)     | NO   |     |         |                |
// | callback             | tinyint(4)       | NO   |     | 0       |                |
// | before_callback_sms  | tinyint(4)       | NO   |     | 0       |                |
// | before_callback_mail | tinyint(4)       | NO   |     | 0       |                |
// | after_callback_sms   | tinyint(4)       | NO   |     | 0       |                |
// | after_callback_mail  | tinyint(4)       | NO   |     | 0  		  |								 |
////////////////////////////////////////////////////////////////////////////////////
type Link struct {
	ID      int64  `json:"id" form:"id" gorm:"column:id"`
	Path    string `json:"path" form:"path" gorm:"column:path"`
	Content string `json:"content" form:"content" gorm:"column:content"`
}

func (this Link) TableName() string {
	return "alert_link"
}
