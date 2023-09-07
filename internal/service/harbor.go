package service

import (
	"context"
	"errors"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gclient"
	"github.com/gogf/gf/v2/util/gconv"
	"publish/internal/model"
)

var (
	clientInfo *model.ClientInfo
)

type sHarbor struct{}

func Harbor() *sHarbor {
	return &sHarbor{}
}

// newHarbor 获取 harbor 客户端
func newHarbor() *model.ClientInfo {
	url := "http://" + AuthArgs.IP + "/api/v2.0"
	c := g.Client()
	c.SetBasicAuth(AuthArgs.Username, AuthArgs.Password)
	clientInfo = &model.ClientInfo{
		Client: c,
		Url:    url,
	}

	return clientInfo
}

func (s *sHarbor) Client() *gclient.Client {
	return clientInfo.Client
}

func (s *sHarbor) Url() string {
	return clientInfo.Url
}

func HarborSetUp() {
	newHarbor()
	g.Log().Info(context.TODO(), "harbor client init success")
}

// ifExistProject check if project exist
// args: project name
func (s *sHarbor) ifExistProject(ctx context.Context, name string) (bool, error) {
	url := s.Url() + "/projects?name=" + name
	r, err := s.Client().Get(ctx, url, g.Map{"with_detail": false})
	if err != nil {
		g.Log().Error(ctx, err)
		return false, err
	}
	defer r.Close()

	res := r.ReadAll()
	if gconv.String(res) == "[]\n" {
		g.Log().Infof(ctx, "project %s not exist", name)
		return false, nil
	}

	return true, err
}

// CreateProject create harbor project
// args: project name
func (s *sHarbor) CreateProject(ctx context.Context, name string) error {
	if ok, _ := s.ifExistProject(ctx, name); !ok {
		url := s.Url() + "/projects"
		// @todo: 需要测试这里是否需要加ContentJson()， res的值为多少
		r, err := s.Client().ContentJson().Post(ctx, url, g.Map{"public": false, "project_name": name})
		if err != nil {
			g.Log().Error(ctx, err)
			return err
		}
		defer r.Close()

		res := r.ReadAll()
		if gconv.String(res) == "" {
			g.Log().Infof(ctx, "create project %s success", name)
			return nil
		} else {
			g.Log().Error(ctx, "create project failed")
			return errors.New("create project failed")
		}
	}

	return nil
}

func (s *sHarbor) ListProjects(ctx context.Context) ([]model.ProjectInfo, error) {
	url := s.Url() + "/projects"
	r, err := s.Client().Get(ctx, url, g.Map{"with_detail": false})
	if err != nil {
		g.Log().Error(ctx, err)
		return nil, err
	}
	defer r.Close()

	res := r.ReadAll()
	var projects []model.ProjectInfo
	err = gjson.Unmarshal(res, &projects)
	if err != nil {
		return nil, err
	}
	g.Log().Debug(ctx, "list projects: %s", projects)
	g.Log().Info(ctx, "list projects success")

	return projects, nil
}

// ListRepos list harbor repos
// args: project name
func (s *sHarbor) ListRepos(ctx context.Context, name, sort string, pSize, p int64) ([]model.RepoInfo, error) {
	url := s.Url() + "/projects/" + name + "/repositories" + "?page=" + gconv.String(p) +
		"&page_size=" + gconv.String(pSize) + "&sort=" + sort
	r, err := s.Client().Get(ctx, url, g.Map{"with_tag": false})
	if err != nil {
		g.Log().Error(ctx, err)
		return nil, err
	}
	defer r.Close()

	res := r.ReadAll()

	var repos []model.RepoInfo
	err = gjson.Unmarshal(res, &repos)

	g.Log().Debug(ctx, "list repos: %s", repos)
	g.Log().Info(ctx, "list repos success")

	return repos, nil
}

// ListDigest list harbor repo's artifact's digest
// args: pName: project name, rName: repo name, sort: sort by, pSize: page size, p: page
func (s *sHarbor) ListDigest(ctx context.Context, pName, rName, sort string, pSize, p int64) ([]model.Artifact, error) {
	url := s.Url() + "/projects/" + pName + "/repositories/" +
		rName + "/artifacts" + "?page=" + gconv.String(p) +
		"&page_size=" + gconv.String(pSize) + "&sort=" + sort
	r, err := s.Client().Get(ctx, url, g.Map{})
	if err != nil {
		g.Log().Error(ctx, err)
		return nil, err
	}
	defer r.Close()

	res := r.ReadAll()
	var artifacts []model.Artifact

	err = gjson.Unmarshal(res, &artifacts)
	if err != nil {
		g.Log().Error(ctx, err)
		return nil, err
	}

	g.Log().Debug(ctx, "list tags page %d and %d items: %s", p, pSize, artifacts)
	g.Log().Info(ctx, "list tags success")

	return artifacts, nil
}

// DeleteTag delete harbor repo's tag
func (s *sHarbor) DeleteTag(ctx context.Context, pName, rName, digest string) error {
	url := s.Url() + "/projects/" + pName + "/repositories/" + rName + "/artifacts/" + digest
	r, err := s.Client().Delete(ctx, url, g.Map{})
	if err != nil {
		g.Log().Error(ctx, err)
		return err
	}
	defer r.Close()
	return err
}

// CreateGcCron create harbor gc schedule
func (s *sHarbor) CreateGcCron(ctx context.Context) error {
	url := s.Url() + "/system/gc/schedule"
	// get cron from config
	cronJob, _ := g.Config().Get(ctx, "harbor.cron")

	scheduleCron := &model.Schedule{
		Cron: cronJob.String(),
		Type: "Custom",
	}
	r, err := s.Client().ContentJson().Post(ctx, url, g.Map{"schedule": scheduleCron})
	if err != nil {
		g.Log().Error(ctx, err)
		return err
	}
	defer r.Close()

	res := r.ReadAll()
	if gconv.String(res) == "" {
		g.Log().Info(ctx, "create gc cron success")
		return nil
	} else {
		g.Log().Error(ctx, "create gc cron failed")
		return errors.New("create gc cron failed")
	}

}

// GetGcCron get schedule of gc job
func (s *sHarbor) GetGcCron(ctx context.Context) (cron interface{}, err error) {
	url := s.Url() + "/system/gc/schedule"
	r, err := s.Client().Get(ctx, url, g.Map{})
	if err != nil {
		g.Log().Error(ctx, err)
		return nil, err
	}
	defer r.Close()

	res := r.ReadAll()
	if string(res) != "" {
		var schedule model.GCSchedule
		err = gjson.Unmarshal(res, &schedule)
		if err != nil {
			g.Log().Error(ctx, err)
			return nil, err
		}

		return schedule, nil
	}

	return
}
