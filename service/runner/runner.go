package runner

import (
	"fmt"
	"github.com/duck8823/minimal-ci/infrastructure/archive/tar"
	"github.com/duck8823/minimal-ci/infrastructure/context"
	"github.com/duck8823/minimal-ci/infrastructure/docker"
	"github.com/duck8823/minimal-ci/infrastructure/logger"
	"github.com/duck8823/minimal-ci/service/github"
	"github.com/pkg/errors"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"os"
	"path"
	"strconv"
	"time"
)

type Runner interface {
	Run(ctx context.Context, repo github.Repository, ref string, command ...string) (plumbing.Hash, error)
	ConvertPullRequestToRef(ctx context.Context, repo github.Repository, num int) (string, error)
}

const NAME = "minimal-ci"

type runnerImpl struct {
	GitHub      github.Service
	Docker      *docker.Client
	Name        string
	BaseWorkDir string
}

func NewWithEnv() (*runnerImpl, error) {
	githubService, err := github.New(os.Getenv("GITHUB_API_TOKEN"))
	if err != nil {
		return nil, errors.WithStack(err)
	}

	dockerClient, err := docker.New()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &runnerImpl{
		GitHub:      githubService,
		Docker:      dockerClient,
		Name:        NAME,
		BaseWorkDir: path.Join(os.TempDir(), NAME),
	}, nil
}

func (r *runnerImpl) ConvertPullRequestToRef(ctx context.Context, repo github.Repository, num int) (string, error) {
	pr, err := r.GitHub.GetPullRequest(ctx, repo, num)
	if err != nil {
		return "", errors.WithStack(err)
	}
	return fmt.Sprintf("refs/heads/%s", pr.GetHead().GetRef()), nil
}

func (r *runnerImpl) Run(ctx context.Context, repo github.Repository, ref string, command ...string) (plumbing.Hash, error) {
	commitHash := make(chan plumbing.Hash, 1)
	errs := make(chan error, 1)

	timeout, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	go func() {
		hash, err := r.run(ctx, repo, ref, command...)
		commitHash <- hash
		errs <- err
	}()

	select {
	case <-timeout.Done():
		hash := <-commitHash
		if timeout.Err() != nil {
			logger.Errorf(ctx.UUID(), "%+v", timeout.Err())
			r.CreateCommitStatusWithError(ctx, repo, <-commitHash, timeout.Err())
		}
		return hash, timeout.Err()
	case err := <-errs:
		hash := <-commitHash
		if err == docker.Failure {
			logger.Error(ctx.UUID(), err.Error())
			r.CreateCommitStatus(ctx, repo, hash, github.FAILURE)
		} else if err != nil {
			logger.Errorf(ctx.UUID(), "%+v", err)
			r.CreateCommitStatusWithError(ctx, repo, hash, err)
		} else {
			r.CreateCommitStatus(ctx, repo, hash, github.SUCCESS)
		}
		return hash, err
	}
}

func (r *runnerImpl) run(ctx context.Context, repo github.Repository, ref string, command ...string) (plumbing.Hash, error) {
	workDir := path.Join(r.BaseWorkDir, strconv.FormatInt(time.Now().Unix(), 10))
	tagName := repo.GetFullName()

	head, err := r.GitHub.Clone(ctx, workDir, repo, ref)
	if err != nil {
		return plumbing.Hash{}, errors.WithStack(err)
	}

	r.CreateCommitStatus(ctx, repo, head, github.PENDING)

	tarFilePath := path.Join(workDir, "minimal-ci.tar")
	writeFile, err := os.OpenFile(tarFilePath, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return head, errors.WithStack(err)
	}
	defer writeFile.Close()

	if err := tar.Create(workDir, writeFile); err != nil {
		return head, errors.WithStack(err)
	}

	readFile, _ := os.Open(tarFilePath)
	defer readFile.Close()

	if err := r.Docker.Build(ctx, readFile, tagName); err != nil {
		return head, errors.WithStack(err)
	}

	if _, err = r.Docker.Run(ctx, docker.Environments{}, tagName, command...); err != nil {
		return head, errors.WithStack(err)
	}

	return head, nil
}

func (r *runnerImpl) CreateCommitStatus(ctx context.Context, repo github.Repository, hash plumbing.Hash, state github.State) {
	msg := fmt.Sprintf("task %s", state)
	if err := r.GitHub.CreateCommitStatus(ctx, repo, hash, &github.Status{
		Context:     &r.Name,
		Description: &msg,
		State:       &state,
	}); err != nil {
		logger.Errorf(ctx.UUID(), "Failed to create commit status: %+v", err)
	}
}

func (r *runnerImpl) CreateCommitStatusWithError(ctx context.Context, repo github.Repository, hash plumbing.Hash, err error) {
	msg := err.Error()
	if len(msg) >= 50 {
		msg = string([]rune(msg)[:46]) + "..."
	}
	state := github.ERROR
	if err := r.GitHub.CreateCommitStatus(ctx, repo, hash, &github.Status{
		Context:     &r.Name,
		Description: &msg,
		State:       &state,
	}); err != nil {
		logger.Errorf(ctx.UUID(), "Failed to create commit status: %+v", err)
	}
}