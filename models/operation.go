package models

type Operation struct {
	ID                 uint   `gorm:"primarykey" json:"id"`
	Namespace          string `gorm:"type:varchar(255);not null" json:"namespace"`
	Action             string `gorm:"type:varchar(255);not null" json:"action"`
	ActionUserName     string `gorm:"type:varchar(255);not null" json:"action_user_name"`
	ActionTime         string `gorm:"type:varchar(255);not null" json:"action_time"`
	ActionTimestamp    string `gorm:"type:varchar(255);not null" json:"action_timestamp"`
	OperatUserName     string `gorm:"type:varchar(255)" json:"operat_user_name"`
	OperationTime      string `gorm:"type:varchar(255)" json:"operation_time"`
	OperationTimestamp string `gorm:"type:varchar(255)" json:"operation_timestamp"`
	GitUrl             string `gorm:"type:varchar(255)" json:"git_url"`
	LastGitBranch      string `gorm:"type:varchar(255)" json:"last_git_branch"`
}
