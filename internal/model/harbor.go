package model

import "github.com/gogf/gf/v2/net/gclient"

type AuthArgs struct {
	IP       string `json:"ip"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type ClientInfo struct {
	Client *gclient.Client `json:"client"`
	Url    string          `json:"url"`
}

type RepoInfo struct {
	Name          string `json:"name"`
	ArtifactCount int64  `json:"artifact_count"`
}

type Artifact struct {
	Digest string        `json:"digest"`
	Tags   []ArtifactTag `json:"tags"`
}

type ArtifactTag struct { // below Artifact's tag list info
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	PullTime     string `json:"pull_time"`
	PushTime     string `json:"push_time"`
	RepositoryID int64  `json:"repository_id"`
}

type ProjectInfo struct {
	Name       string `json:"name"`
	RepoCount  int64  `json:"repo_count"`
	CreateTime string `json:"creation_time"`
}

type Schedule struct {
	Cron             string `json:"cron"`
	Type             string `json:"type"`
	NextScheduleTime string `json:"next_schedule_time"`
}

type GCSchedule struct {
	Schedule Schedule `json:"schedule"`
}
