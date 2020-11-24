package changelog

import (
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"yunion.io/x/pkg/errors"

	"github.com/yunionio/git-tools/pkg/gitlib"
	"github.com/yunionio/git-tools/pkg/types"
)

type GlobalGenerator struct {
	config    *types.GlobalChangeLogConfig
	processor gitlib.Processor
}

// NewGlobalGenerator create new GlobalGenerator
func NewGlobalGenerator(config *types.GlobalChangeLogConfig) *GlobalGenerator {
	return &GlobalGenerator{
		config: config,
	}
}

func (gen *GlobalGenerator) GetResults() (*types.GlobalChangeLogResult, error) {
	ret := &types.GlobalChangeLogResult{
		Releases: make([]*types.ReleaseChangeLogResult, len(gen.config.Releases)),
	}

	for idx, rls := range gen.config.Releases {
		rRet, err := gen.GetReleaseResults(rls)
		if err != nil {
			return nil, errors.Wrapf(err, "get release results")
		}

		ret.Releases[idx] = rRet
	}

	return ret, nil
}

func (gen *GlobalGenerator) getProcesser(repo *types.Repository) gitlib.Processor {
	// TODO: support others
	return &gitlib.GitHubProcessor{}
}

func GetBranchWeight(branch string) (int, error) {
	verStr, err := GetSemverBranchVersion(branch)
	if err != nil {
		return 0, errors.Wrapf(err, "GetSemverBranchVersion %q", branch)
	}
	return GetSemverStrWeight(verStr)
}

func GetSemverStrWeight(verStr string) (int, error) {
	verStr = strings.ReplaceAll(verStr, ".", "")
	weight, err := strconv.Atoi(verStr)
	if err != nil {
		return 0, errors.Wrapf(err, "parse sem version string %q", verStr)
	}
	return weight, nil
}

func (gen *GlobalGenerator) GetReleaseResults(rls *types.ReleaseChangeLogConfig) (*types.ReleaseChangeLogResult, error) {
	ret := &types.ReleaseChangeLogResult{
		Branch: rls.Branch,
		Repos:  make([]*types.RepoChangelogResult, len(rls.Repos)),
	}
	branchWeight, err := GetBranchWeight(rls.Branch)
	if err != nil {
		return nil, errors.Wrapf(err, "GetBranchWeight %q", rls.Branch)
	}
	ret.Weight = branchWeight

	for idx := range rls.Repos {
		repo := rls.Repos[idx]
		conf := rls.ToChangelogConfig(gen.config.Bin, gen.config.Options, idx)

		gen := NewGenerator(conf, gen.getProcesser(repo))
		unreleased, versions, err := gen.GetSemverBranchResults(rls.Branch)
		if err != nil {
			return nil, errors.Wrapf(err, "get results for branch %q", rls.Branch)
		}
		repoRet := &types.RepoChangelogResult{
			Repo:       repo,
			Versions:   versions,
			Unreleased: unreleased,
		}
		ret.Repos[idx] = repoRet
	}

	return ret, nil
}

func (gen *GlobalGenerator) GetRenderData() (*types.GlobalRenderData, error) {
	results, err := gen.GetResults()
	if err != nil {
		return nil, errors.Wrap(err, "get results")
	}

	data, err := NewGlobalRenderData(results)
	if err != nil {
		return nil, errors.Wrap(err, "results to render data")
	}

	return data, nil
}

func (gen *GlobalGenerator) Generate(w io.Writer) error {
	data, err := gen.GetRenderData()
	if err != nil {
		return err
	}
	return gen.render(w, data)
}

func (gen *GlobalGenerator) render(w io.Writer, results *types.GlobalRenderData) error {
	if _, err := os.Stat(gen.config.Template); err != nil {
		return err
	}

	fname := filepath.Base(gen.config.Template)

	t := template.Must(template.New(fname).Funcs(TemplateFuncMap).ParseFiles(gen.config.Template))

	return t.Execute(w, results)
}
