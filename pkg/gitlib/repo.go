package gitlib

import (
	"fmt"
	"os"

	"yunion.io/x/log"
	"yunion.io/x/pkg/errors"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
)

type Repository struct {
	localDir string
	url      string
	*git.Repository
}

func NewRepository(localDir string, repoURL string) (*Repository, error) {
	_, err := os.Stat(localDir)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}

		if err := os.MkdirAll(localDir, os.ModePerm); err != nil {
			return nil, errors.Wrapf(err, "MkdirAll %q", localDir)
		}

		// local repo not exist, clone it
		log.Infof("start clone %q to %q", repoURL, localDir)
		repo, err := git.PlainClone(localDir, false, &git.CloneOptions{
			URL:      repoURL,
			Progress: os.Stdout,
		})
		if err != nil {
			return nil, errors.Wrapf(err, "clone %q to local %q", repoURL, localDir)
		}
		return newRepository(repo, localDir, repoURL), nil
	}

	// local repo already exist, open it
	repo, err := git.PlainOpen(localDir)
	if err != nil {
		return nil, errors.Wrapf(err, "open local repo %q", localDir)
	}

	return newRepository(repo, localDir, repoURL), nil
}

func newRepository(repo *git.Repository, localDir string, repoURL string) *Repository {
	return &Repository{
		Repository: repo,
		localDir:   localDir,
		url:        repoURL,
	}
}

func (repo *Repository) GetConfig() (*config.Config, error) {
	return repo.Repository.Config()
}

func (repo *Repository) GetOriginRemote() (*config.RemoteConfig, error) {
	config, err := repo.GetConfig()
	if err != nil {
		return nil, errors.Wrap(err, "get config")
	}

	remote, ok := config.Remotes["origin"]
	if !ok {
		return nil, errors.Error("not found 'origin' remote")
	}

	return remote, nil
}

func (repo *Repository) GetURL() (string, error) {
	origin, err := repo.GetOriginRemote()
	if err != nil {
		return "", errors.Wrap(err, "GetOriginRemote")
	}

	return origin.URLs[0], nil
}

func (repo *Repository) LogPrefix() string {
	return fmt.Sprintf("%s: %s", repo.localDir, repo.url)
}

func (repo *Repository) Fetch() error {
	log.Infof("start fetch %q", repo.LogPrefix())

	err := repo.Repository.Fetch(&git.FetchOptions{
		Tags:     git.AllTags,
		Progress: os.Stdout,
	})
	if err == git.NoErrAlreadyUpToDate {
		return nil
	}

	return err
}
