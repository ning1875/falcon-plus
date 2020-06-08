package model

type PolyItem struct {
	PolyName string  `json:"poly_name"`
	Value    float64 `json:"value"`
}

type PolyRequest struct {
	PolyName string      `json:"poly_name"`
	EndPoint string      `json:"end_point"`
	Type     string      `json:"counterType"`
	Value    interface{} `json:"value"`
}
