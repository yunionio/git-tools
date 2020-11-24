package run

import (
	"fmt"
	"io/ioutil"
	"path"
	"strings"

	"github.com/spf13/cobra"

	"yunion.io/x/jsonutils"
	"yunion.io/x/pkg/errors"

	"github.com/yunionio/git-tools/pkg/changelog"
	"github.com/yunionio/git-tools/pkg/gitlib"
	"github.com/yunionio/git-tools/pkg/types"
)

var (
	Cmd = &cobra.Command{
		Use:   "run",
		Short: "Generate changelog",
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(configFile)
		},
	}
)

var (
	configFile   string
	noFetch      bool
	outputFormat string
)

func init() {
	Cmd.Flags().StringVarP(&configFile, "config", "c", "", "Config file (required)")
	Cmd.MarkFlagRequired("config")
	Cmd.Flags().BoolVarP(&noFetch, "no-fetch", "n", false, "Not fetch each repository")
	Cmd.Flags().StringVarP(&outputFormat, "output-format", "o", "", "Output format for raw render data, choices(`json|yaml`)")
}

func initLocalRepos(config *types.GlobalChangeLogConfig) error {
	for _, rls := range config.Releases {
		for _, repo := range rls.Repos {
			// set repo default name
			repo.URL = strings.TrimRight(repo.URL, "/")
			urlSegs := strings.Split(repo.URL, "/")
			if len(urlSegs) == 0 {
				return errors.Errorf("Invalid repo url %q", repo.URL)
			}
			if repo.Name == "" {
				repo.Name = urlSegs[len(urlSegs)-1]
			}
			if repo.WorkingDir == "" {
				repo.WorkingDir = path.Join(config.CacheDir, repo.Name)
			}

			repoObj, err := gitlib.NewRepository(repo.WorkingDir, repo.URL)
			if err != nil {
				return errors.Wrapf(err, "newRepository %q", repo.URL)
			}

			if noFetch {
				return nil
			}
			if err := repoObj.Fetch(); err != nil {
				return errors.Wrapf(err, "fetch repo %s", repoObj.LogPrefix())
			}
		}
	}

	return nil
}

func normalizeConfig(config *types.GlobalChangeLogConfig) {
	if config.Output == nil {
		config.Output = &types.GlobalChangelogOutConfig{
			Dir: "./_output/changelog",
		}
	}

	if config.Options == nil {
		config.Options = new(types.ChangelogConfigOptions)
	}

	opt := config.Options
	opt.UseSemVer = true
	opt.NoMerges = true
	if opt.CommitGroupTitleMaps == nil {
		opt.CommitGroupTitleMaps = make(map[string]string)
	}

	for key, title := range map[string]string{
		"feat":     "Features",
		"fix":      "Bug Fixes",
		"perf":     "Performance Improvements",
		"refactor": "Code Refactoring",
	} {
		opt.CommitGroupTitleMaps[key] = title
	}

	if len(opt.HeaderPatternMaps) == 0 {
		opt.HeaderPatternMaps = []string{"Type", "Scope", "Subject"}
	}
	if len(opt.HeaderPattern) == 0 {
		opt.HeaderPattern = "^(\\w*)(?:\\(([\\w\\$\\.\\-\\*\\s]*)\\))?\\:\\s(.*)$"
	}
	if opt.CommitGroupBy == "" {
		opt.CommitGroupBy = "Type"
	}
	if opt.CommitGroupSortBy == "" {
		opt.CommitGroupSortBy = "Title"
	}
	if opt.CommitSortBy == "" {
		opt.CommitSortBy = "Scope"
	}
	if len(opt.NoteKeywords) == 0 {
		opt.NoteKeywords = []string{"BREAKING CHANGE"}
	}
}

func run(configFile string) error {
	content, err := ioutil.ReadFile(configFile)
	if err != nil {
		return errors.Wrapf(err, "read config file %q", configFile)
	}

	jObj, err := jsonutils.ParseYAML(string(content))
	if err != nil {
		return errors.Wrapf(err, "parse config %s yaml content", configFile)
	}

	configV1 := new(types.GlobalChangeLogConfigV1)
	if err := jObj.Unmarshal(configV1); err != nil {
		return errors.Wrap(err, "load config")
	}
	config, err := configV1.ToInternalConfig()
	if err != nil {
		return errors.Wrap(err, "config to internal config")
	}
	normalizeConfig(config)

	if err := initLocalRepos(config); err != nil {
		return errors.Wrap(err, "init local repository")
	}

	gen := changelog.NewGlobalGenerator(config)
	result, err := gen.GetRenderData()
	if err != nil {
		return errors.Wrap(err, "generate render data")
	}

	return processData(result, outputFormat, config.Template, config.Output)
}

func processData(data *types.GlobalRenderData, outputFormat string, templateFile string, config *types.GlobalChangelogOutConfig) error {
	if outputFormat != "" {
		obj := jsonutils.Marshal(data)
		switch outputFormat {
		case "json":
			fmt.Printf(obj.PrettyString())
		case "yaml":
			fmt.Printf(obj.YAMLString())
		default:
			return errors.Errorf("Not support output format: %q", outputFormat)
		}
		return nil
	}

	return handleOutput(data, templateFile, config)
}
