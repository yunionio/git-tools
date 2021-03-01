package changelog

import (
	"fmt"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/blang/semver/v4"

	"yunion.io/x/pkg/errors"

	"github.com/yunionio/git-tools/pkg/types"
)

var (
	TemplateFuncMap template.FuncMap
)

func init() {
	tagNameRef := func(repoName string, tag *types.Tag) string {
		ref := fmt.Sprintf("%s - %s", repoName, tag.Name)
		if tag.Previous != nil {
			ref = fmt.Sprintf("[%s]", ref)
		}
		return ref
	}

	TemplateFuncMap = template.FuncMap{
		// format the input time according to layout
		"datetime": func(layout string, input time.Time) string {
			return input.Format(layout)
		},
		// check whether substs is withing s
		"contains": func(s, substr string) bool {
			return strings.Contains(s, substr)
		},
		// check whether s begins with prefix
		"hasPrefix": func(s, prefix string) bool {
			return strings.HasPrefix(s, prefix)
		},
		// check whether s ends with suffix
		"hasSuffix": func(s, suffix string) bool {
			return strings.HasSuffix(s, suffix)
		},
		// replace the first n instances of old with new
		"replace": func(s, old, new string, n int) string {
			return strings.Replace(s, old, new, n)
		},
		// lower case a string
		"lower": func(s string) string {
			return strings.ToLower(s)
		},
		// upper case a string
		"upper": func(s string) string {
			return strings.ToUpper(s)
		},
		// upper case the first character of a string
		"upperFirst": func(s string) string {
			if len(s) > 0 {
				return strings.ToUpper(string(s[0])) + s[1:]
			}
			return ""
		},
		"tagNameRef": tagNameRef,
		// tagNameDate get the tag name with date
		"tagNameDate": func(repoName string, tag *types.Tag) string {
			tagName := tagNameRef(repoName, tag)
			dateStr := tag.Date.Format("2006-01-02")
			tagName = fmt.Sprintf("%s - %s", tagName, dateStr)
			return tagName
		},
		// tagRef get the tag reference link url
		"tagRef": func(tag *types.Tag, repoName string, repoURL string) string {
			if tag.Previous == nil {
				return fmt.Sprintf("[%s]: %s/tree/%s", tagNameRef(repoName, tag), repoURL, tag.Name)
			}
			return fmt.Sprintf("%s: %s/compare/%s...%s", tagNameRef(repoName, tag), repoURL, tag.Previous.Name, tag.Name)
		},
		// commitSummary get the commit summary string
		"commitSummary": templateCommitSummary,
		// isCommitsEmpty
		"isCommitsNotEmpty": func(commits []*types.Commit) bool {
			return len(commits) != 0
		},
	}
}

func templateCommitSummary(commit *types.Commit) string {
	var summary string

	scope := commit.Scope
	if scope != "" {
		summary = fmt.Sprintf("**%s:** ", scope)
	}
	if commit.Subject != "" {
		summary = fmt.Sprintf("%s%s", summary, commit.Subject)
	} else {
		summary = fmt.Sprintf("%s%s", summary, commit.Header)
	}

	summary = fmt.Sprintf("%s (%s, [%s](mailto:%s))", summary, commit.Hash.Short, commit.Author.Name, commit.Author.Email)
	return summary
}

func NewGlobalRenderData(result *types.GlobalChangeLogResult) (*types.GlobalRenderData, error) {
	data := &types.GlobalRenderData{
		Releases: make([]*types.ReleaseRenderData, len(result.Releases)),
	}

	for idx := range result.Releases {
		var err error

		rls := result.Releases[idx]
		data.Releases[idx], err = NewReleaseRenderData(rls)
		if err != nil {
			return nil, err
		}
	}

	return data, nil
}

func getGlobalVersionRenderDatas(input map[string]*types.GlobalVersionRenderData) ([]*types.GlobalVersionRenderData, error) {
	tagVers := make([]*semver.Version, 0)
	for verStr := range input {
		v, err := semver.Parse(verStr)
		if err != nil {
			return nil, errors.Wrapf(err, "parse version %q", verStr)
		}
		tagVers = append(tagVers, &v)
	}

	sort.Slice(tagVers, func(i, j int) bool {
		return tagVers[i].GE(*tagVers[j])
	})

	ret := make([]*types.GlobalVersionRenderData, 0)

	for _, ver := range tagVers {
		repos := input[ver.String()]
		repos.Sort()
		ret = append(ret, repos)
	}

	return ret, nil
}

func NewReleaseRenderData(rls *types.ReleaseChangeLogResult) (*types.ReleaseRenderData, error) {
	data := &types.ReleaseRenderData{
		Branch:   rls.Branch,
		Weight:   rls.Weight,
		Versions: make([]*types.GlobalVersionRenderData, 0),
	}

	versionMap := make(map[string]*types.GlobalVersionRenderData, 0)

	for _, repo := range rls.Repos {
		for _, version := range repo.Versions {
			tagVer := version.Tag.Version
			tagVerStr := tagVer.String()
			group, ok := versionMap[tagVerStr]
			if !ok {
				tagWeight, err := GetSemverStrWeight(tagVerStr)
				if err != nil {
					return nil, errors.Wrapf(err, "GetSemverStrWeight %q, repo %q", tagVerStr, repo.Repo.Name)
				}
				versionMap[tagVerStr] = &types.GlobalVersionRenderData{
					TagName: tagVerStr,
					Weight:  tagWeight,
					Date:    version.Tag.Date,
					Repos: []*types.RepoVersionRenderData{
						newRepoVersionRenderData(repo.Repo, version),
					},
				}
			} else {
				group.Repos = append(group.Repos, newRepoVersionRenderData(repo.Repo, version))
			}
		}
	}

	sortVersions, err := getGlobalVersionRenderDatas(versionMap)
	if err != nil {
		return nil, errors.Wrap(err, "getGlobalVersionRenderDatas")
	}

	for _, item := range sortVersions {
		data.Versions = append(data.Versions, item)
	}

	return data, nil
}

func newRepoVersionRenderData(repo *types.Repository, version *types.Version) *types.RepoVersionRenderData {
	return &types.RepoVersionRenderData{
		Repo:    repo,
		Version: version,
	}
}
