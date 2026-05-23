package config

// updateStatusType 更新状态
type UpdateStatusType string

const (
	UpdatedNothing UpdateStatusType = "未改变"
	UpdatedFailed  UpdateStatusType = "失败"
	UpdatedSuccess UpdateStatusType = "成功"
)
