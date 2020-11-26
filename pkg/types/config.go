package types

import (
	"path"
	"strings"

	"yunion.io/x/pkg/errors"
)

var (
	ExampleGlocalChangeLogConfigv1 *GlobalChangeLogConfigV1 = &GlobalChangeLogConfigV1{
		Template: "./template/CHANGELOG.tpl.md",
		CacheDir: "./_cache/",
		Output: &GlobalChangelogOutConfig{
			Dir: "./_output/changelog",
		},
		Options: &ChangelogConfigOptions{
			UseSemVer: true,
			NoMerges:  true,
			CommitGroupTitleMaps: map[string]string{
				"feat":     "Features",
				"fix":      "Bug Fixes",
				"perf":     "Performance Improvements",
				"refactor": "Code Refactoring",
			},
			HeaderPatternMaps: []string{"Type", "Scope", "Subject"},
			HeaderPattern:     `^(\w*)(?:\(([\w\$\.,\/\-\*\s]*)\))?\:\s(.*)$`,
			CommitGroupBy:     "Type",
			CommitGroupSortBy: "Title",
			CommitSortBy:      "Scope",
			NoteKeywords:      []string{"BREAKING CHANGE"},
		},
		Releases: []*ReleaseChangeLogConfigV1{
			{
				Branch: "release/3.4",
				Repos: []string{
					"https://github.com/yunionio/notify-plugins",
					"https://github.com/yunionio/onecloud-service-operator",
					"https://github.com/yunionio/ocadm",
					"https://github.com/yunionio/onecloud",
					"https://github.com/yunionio/onecloud-operator",
					"https://github.com/yunionio/sdnagent",
				},
			},
			{
				Branch: "release/3.3",
				Repos: []string{
					"https://github.com/yunionio/notify-plugins",
					"https://github.com/yunionio/onecloud-service-operator",
					"https://github.com/yunionio/ocadm",
					"https://github.com/yunionio/onecloud",
					"https://github.com/yunionio/onecloud-operator",
					"https://github.com/yunionio/sdnagent",
				},
			},
		},
	}
)

type GlobalChangeLogConfigV1 struct {
	// Path for template file
	Template string
	// CacheDir for local repository clone directory
	CacheDir string `json:"cacheDir"`
	// Options configure generate changelog options
	Options *ChangelogConfigOptions `json:"options"`
	// Releases is each release branch want to generate changelog
	Releases []*ReleaseChangeLogConfigV1 `json:"releases"`
	// Output configure output handle options
	Output *GlobalChangelogOutConfig `json:"output"`
}

func (c *GlobalChangeLogConfigV1) ToInternalConfig() (*GlobalChangeLogConfig, error) {
	if c.CacheDir == "" {
		return nil, errors.Errorf("cacheDir must specified")
	}

	ic := &GlobalChangeLogConfig{
		Bin:      "git",
		CacheDir: c.CacheDir,
		Template: c.Template,
		Options:  c.Options,
		Output:   c.Output,
	}

	for _, rls := range c.Releases {
		iRls, err := rls.ToInternalConfig(c.CacheDir)
		if err != nil {
			return nil, errors.Wrapf(err, "release config")
		}

		ic.Releases = append(ic.Releases, iRls)
	}

	return ic, nil
}

type ReleaseChangeLogConfigV1 struct {
	Branch string   `json:"branch"`
	Repos  []string `json:"repos"`
}

func (c *ReleaseChangeLogConfigV1) ToInternalConfig(cacheDir string) (*ReleaseChangeLogConfig, error) {
	ic := &ReleaseChangeLogConfig{
		Branch: c.Branch,
	}

	for _, url := range c.Repos {
		repo := new(Repository)
		repo.URL = strings.TrimRight(url, "/")
		urlSegs := strings.Split(url, "/")
		if len(urlSegs) == 0 {
			return nil, errors.Errorf("Invalid repo url %q", repo.URL)
		}
		repo.Name = urlSegs[len(urlSegs)-1]
		repo.WorkingDir = path.Join(cacheDir, repo.Name)
		ic.Repos = append(ic.Repos, repo)
	}

	return ic, nil
}
