package models

type NormalGCPercent struct {
	Value int64 `json:"value" form:"value" binding:"Required"`
}

type StartupGCPercent struct {
	Value int64 `json:"value" form:"value" binding:"Required"`
}
