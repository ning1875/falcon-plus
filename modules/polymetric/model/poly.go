package model

type PolyMetric struct {
	Id         int    `json:"id" gorm:"id"`
	PolyType   string `json:"poly_type" gorm:"poly_type"`
	Name       string `json:"name" gorm:"name"`
	CreateUser string `json:"create_user" gorm:"create_user"`
	Counter    string `json:"counter" gorm:"counter"`
	CreateAt   string `json:"create_at" gorm:"create_at"`
}
