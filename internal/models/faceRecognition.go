package models

type Tenant struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Surname  string `json:"surname"`
	Phone    string `json:"phone"`
	DeviceID string `json:"deviceID"`
}
