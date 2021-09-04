package models

type Account struct {
	BaseModel
	Key string `json:"key"`
	Value string `json:"value"`
}
