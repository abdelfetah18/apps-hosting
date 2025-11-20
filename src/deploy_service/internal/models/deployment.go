package models

import "time"

type DeploymentStatus string

const (
	DeploymentStatusPending   DeploymentStatus = "pending"
	DeploymentStatusSuccessed DeploymentStatus = "successed"
	DeploymentStatusFailed    DeploymentStatus = "failed"
)

type Deployment struct {
	Id        string           `bun:"id,pk,type:uuid,default:gen_random_uuid()" json:"id"`
	BuildId   string           `bun:"build_id" json:"build_id"`
	AppId     string           `bun:"app_id" json:"app_id"`
	Status    DeploymentStatus `bun:"status" json:"status"`
	CreatedAt time.Time        `bun:"created_at,default:now()" json:"created_at"`
}
