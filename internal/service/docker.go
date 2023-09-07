package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gfile"
	"io"
	"os"
	"publish/internal/model"
	"publish/utility/docker"
	"strings"
)

var (
	AuthArgs     *model.AuthArgs
	dockerClient *client.Client
	registryAuth string
)

type sDocker struct{}

func Docker() *sDocker {
	return &sDocker{}
}

// init initial harbor auth info
func init() {
	username, _ := g.Config().Get(context.TODO(), "harbor.username")
	password, _ := g.Config().Get(context.TODO(), "harbor.password")
	ip, _ := g.Config().Get(context.TODO(), "harbor.ip")
	AuthArgs = &model.AuthArgs{
		IP:       ip.String(),
		Username: username.String(),
		Password: password.String(),
	}
}

func newDockerClient() *client.Client {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	dockerClient = cli

	return dockerClient
}

func dockerLoginStr() {
	auth := types.AuthConfig{
		ServerAddress: AuthArgs.IP,
		Username:      AuthArgs.Username,
		Password:      AuthArgs.Password,
	}
	authBytes, _ := json.Marshal(auth)
	registryAuth = base64.URLEncoding.EncodeToString(authBytes)
}

func (s *sDocker) Client() *client.Client {
	return dockerClient
}

func DockerSetUP() {
	newDockerClient()
	g.Log().Info(context.TODO(), "docker client init success")
	dockerLoginStr()
	g.Log().Info(context.TODO(), "docker login info init success")
}

func (s *sDocker) pushImageToHarbor(ctx context.Context, image string) error {
	res, err := dockerClient.ImagePush(ctx, image, types.ImagePushOptions{
		RegistryAuth: registryAuth,
	})
	if err != nil {
		g.Log().Error(ctx, err.Error())
	}
	defer res.Close()

	err = docker.PrintDockerMsg(res)
	if err != nil {
		g.Log().Error(ctx, err)
		return err
	}
	g.Log().Infof(ctx, "push image %s to harbor success", image)
	return nil
}

func (s *sDocker) pullImage(ctx context.Context, image string) error {
	res, err := dockerClient.ImagePull(ctx, image, types.ImagePullOptions{
		RegistryAuth: registryAuth,
	})
	if err != nil {
		g.Log().Error(ctx, err.Error())
	}
	defer res.Close()

	err = docker.PrintDockerMsg(res)
	if err != nil {
		g.Log().Error(ctx, err)
		return err
	}
	g.Log().Infof(ctx, "pull image %s from harbor success", image)
	return nil
}

func (s *sDocker) RemoveImage(ctx context.Context, image string) error {
	_, err := dockerClient.ImageRemove(ctx, image, types.ImageRemoveOptions{
		Force: true,
	})
	if err != nil {
		g.Log().Error(ctx, err.Error())
		return err
	}
	g.Log().Infof(ctx, "remove image %s success", image)
	return nil
}

func (s *sDocker) ListImages(ctx context.Context) ([]types.ImageSummary, error) {
	images, err := dockerClient.ImageList(ctx, types.ImageListOptions{})
	if err != nil {
		g.Log().Error(ctx, err.Error())
		return nil, err
	}
	return images, nil
}

// tagImage tag image
// image: old image name
// tag: new image name
func (s *sDocker) tagImage(ctx context.Context, image, tag string) error {
	err := dockerClient.ImageTag(ctx, image, tag)
	if err != nil {
		g.Log().Error(ctx, err.Error())
		return err
	}
	g.Log().Infof(ctx, "tag image %s success", image)
	return nil
}

// addImagePrefix add harbor ip to image name
// if image name is "nginx:latest", then return "ip/nginx:latest"
func (s *sDocker) addImagePrefix(image string) string {
	return AuthArgs.IP + "/" + image
}

// modifyImagePrefix modify image name
// if image name is "otherIp/nginx:latest", then return "harborIp/nginx:latest"
func (s *sDocker) modifyImagePrefix(image string) string {
	return AuthArgs.IP + "/" + s.getImageName(image)
}

func (s *sDocker) getImageName(image string) string {
	return image[strings.Index(image, "/")+1:]
}

func (s *sDocker) getImageTag(image string) string {
	return image[strings.LastIndex(image, ":")+1:]
}

// getImageProject get image project
func (s *sDocker) getImageProject(image string) string {
	return strings.Split(image, "/")[1]
}

func (s *sDocker) TagDockerImage(ctx context.Context, image string) (error, string) {
	var newImage string
	// if image is private image, then modify image prefix
	// else add image prefix
	if ip := docker.FindIpAddress(strings.Split(image, "/")[0]); ip != "" {
		newImage = s.modifyImagePrefix(image)
	} else {
		newImage = s.addImagePrefix(image)
	}
	err := s.tagImage(ctx, image, newImage)
	if err != nil {
		return err, ""
	}
	return nil, newImage
}

func (s *sDocker) PushDockerImage(ctx context.Context, image string) (err error) {
	project := s.getImageProject(image)
	// if project not exist, then create project
	err = Harbor().CreateProject(ctx, project)
	if err != nil {
		return err
	}
	err = s.pushImageToHarbor(ctx, image)
	if err != nil {
		return err
	}
	return nil
}

// PullImageAndSaveToLocal pull image from harbor and save to local
// imageDir: image save dir
// images: image name list
func (s *sDocker) PullImageAndSaveToLocal(ctx context.Context, imageDir string, images []string) (err error) {
	// create images dir
	if err = Path().CreateDir(ctx, imageDir); err != nil {
		return err
	}

	// change dir to images dir
	_ = gfile.Chdir(imageDir)
	for _, image := range images {
		err = s.pullImage(ctx, image)
		if err != nil {
			return err
		}
		// save image to local
		err = s.saveImageTag(ctx, image)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *sDocker) saveImageTag(ctx context.Context, image string) error {
	// save image to local
	jobName := gfile.Basename(strings.Split(image, ":")[0])
	saveTarFile := jobName + ".tar"
	f, err := os.Create(saveTarFile)
	if err != nil {
		g.Log().Error(ctx, err.Error())
		return err
	}
	defer f.Close()

	rd, err := s.Client().ImageSave(ctx, []string{image})
	if err != nil {
		g.Log().Error(ctx, err.Error())
		return err
	}
	defer rd.Close()

	_, err = io.Copy(f, rd)
	if err != nil {
		g.Log().Error(ctx, err.Error())
		return err
	}

	g.Log().Infof(ctx, "%s save as %s", image, saveTarFile)
	return nil
}

// LoadImage load image from local
// image: image tar file
func (s *sDocker) LoadImage(ctx context.Context, image string) (error, string) {
	f, err := os.Open(image)
	if err != nil {
		return err, ""
	}
	defer f.Close()

	res, err := s.Client().ImageLoad(ctx, f, true)
	if err != nil {
		return err, ""
	}
	defer res.Body.Close()

	var stream model.ImageTag
	err = json.NewDecoder(res.Body).Decode(&stream)
	if err != nil {
		return err, ""
	}

	index := strings.Index(stream.Stream, ":")
	result := strings.TrimSpace(stream.Stream[index+1:])

	g.Log().Infof(ctx, "load image %s success", image)
	return nil, result
}
