package models

import "time"

type BuildStatus string

const (
	BuildStatusPending   BuildStatus = "pending"
	BuildStatusSuccessed BuildStatus = "successed"
	BuildStatusFailed    BuildStatus = "failed"
)

type Build struct {
	Id         string      `bun:"id,pk,type:uuid,default:gen_random_uuid()" json:"id"`
	AppId      string      `bun:"app_id" json:"app_id"`
	Status     BuildStatus `bun:"status" json:"status"`
	ImageURL   string      `bun:"image_url" json:"image_url"`
	CommitHash string      `bun:"commit_hash" json:"commit_hash"`
	CreatedAt  time.Time   `bun:"created_at,default:now()" json:"created_at"`
}
