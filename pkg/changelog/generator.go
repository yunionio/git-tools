package changelog

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	gitcmd "github.com/tsuyoshiwada/go-gitcmd"

	"yunion.io/x/pkg/errors"

	"github.com/yunionio/git-tools/pkg/gitlib"
	"github.com/yunionio/git-tools/pkg/types"
)

// Generator of CHANGELOG
type Generator struct {
	client          gitcmd.Client
	config          *types.ChangelogConfig
	tagReader       gitlib.TagReader
	tagSelector     gitlib.TagSelector
	commitParser    gitlib.CommitParser
	commitExtractor gitlib.CommitExtractor
	processor       gitlib.Processor
}

// NewGenerator receives `Config` and create an new `Generator`
func NewGenerator(config *types.ChangelogConfig, processor gitlib.Processor) *Generator {
	cli := gitcmd.New(&gitcmd.Config{
		Bin: config.Bin,
	})

	if processor != nil {
		processor.Bootstrap(config)
	}

	normalizeConfig(config)

	var tagReader gitlib.TagReader
	if !config.Options.UseSemVer {
		tagReader = gitlib.NewTagReader(cli, config.Options.TagFilterPattern)
	} else {
		tagReader = gitlib.NewSemVerTagReader(cli)
	}

	return &Generator{
		client:          cli,
		config:          config,
		tagReader:       tagReader,
		tagSelector:     gitlib.NewTagSelector(),
		commitParser:    gitlib.NewCommitParser(cli, config),
		commitExtractor: gitlib.NewCommitExtractor(config.Options),
		processor:       processor,
	}
}

func normalizeConfig(config *types.ChangelogConfig) {
	opts := config.Options

	if opts.HeaderPattern == "" {
		opts.HeaderPattern = "^(.*)$"
		opts.HeaderPatternMaps = []string{
			"Subject",
		}
	}

	if opts.MergePattern == "" {
		opts.MergePattern = "^Merge branch '(\\w+)'$"
		opts.MergePatternMaps = []string{
			"Source",
		}
	}

	if opts.RevertPattern == "" {
		opts.RevertPattern = "^Revert \"([\\s\\S]*)\"$"
		opts.RevertPatternMaps = []string{
			"Header",
		}
	}

	config.Options = opts
}

func (gen *Generator) workdir() (func() error, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, errors.Wrap(err, "get current workdir")
	}

	if err := os.Chdir(gen.config.WorkingDir); err != nil {
		return nil, errors.Wrapf(err, "chdir to %q", gen.config.WorkingDir)
	}

	return func() error {
		if err := os.Chdir(cwd); err != nil {
			return errors.Wrapf(err, "chdir back to %q", cwd)
		}
		return nil
	}, nil
}

// Generate gets the commit based on the specified tag `query` and writes the result to `io.Writer`
//
// tag `query` can be specified with the following rule
//  <old>..<new> - Commit contained in `<new>` tags from `<old>` (e.g. `1.0.0..2.0.0`)
//  <tagname>..  - Commit from the `<tagname>` to the latest tag (e.g. `1.0.0..`)
//  ..<tagname>  - Commit from the oldest tag to `<tagname>` (e.g. `..1.0.0`)
//  <tagname>    - Commit contained in `<tagname>` (e.g. `1.0.0`)
func (gen *Generator) Generate(w io.Writer, query string) error {
	unreleased, versions, err := gen.GetResults(query)
	if err != nil {
		return err
	}

	return gen.render(w, unreleased, versions)
}

func (gen *Generator) GeneratorBySemverBranch(w io.Writer, branch string) error {
	unreleased, versions, err := gen.GetSemverBranchResults(branch)
	if err != nil {
		return err
	}

	return gen.render(w, unreleased, versions)
}

// GetSemverBranchTags read tags according by branch
// branch format is `release/major.minor`
func (gen *Generator) GetSemverBranchTags(branch string) ([]*types.Tag, error) {
	branchVer, err := GetSemverBranchVersion(branch)
	if err != nil {
		return nil, err
	}

	tags, err := gen.tagReader.ReadAll()
	if err != nil {
		return nil, err
	}

	return filterTagsByPrefix(branchVer, tags), nil
}

func filterTagsByPrefix(branchVer string, tags []*types.Tag) []*types.Tag {
	ret := make([]*types.Tag, 0)

	for _, tag := range tags {
		if strings.HasPrefix(tag.Name, "v"+branchVer) {
			ret = append(ret, tag)
		}
	}

	return ret
}

// GetSemverBranchVersion read branch semantic version string
// branch format is `release/major.minor`
func GetSemverBranchVersion(branch string) (string, error) {
	reRef := regexp.MustCompile(`release[\/-]([\d]+\.[\d]+$)`)
	res := reRef.FindAllStringSubmatch(branch, -1)
	if len(res) == 0 {
		return "", errors.Errorf("branch %q is not release branch", branch)
	}

	branchVer := res[0][1]
	return branchVer, nil
}

func (gen *Generator) GetSemverBranchQuery(branch string) (string, error) {
	back, err := gen.workdir()
	if err != nil {
		return "", err
	}
	defer back()

	tags, err := gen.GetSemverBranchTags(branch)
	if err != nil {
		return "", err
	}

	if len(tags) == 0 {
		return "", errors.Errorf("branch %q not found tags", branch)
	}

	query := tags[0].Name
	if len(tags) > 1 {
		query = fmt.Sprintf("%s..%s", tags[len(tags)-1].Name, tags[0].Name)
	}

	return query, nil
}

func (gen *Generator) GetSemverBranchResults(branch string) (*types.Unreleased, []*types.Version, error) {
	query, err := gen.GetSemverBranchQuery(branch)
	if err != nil {
		return nil, nil, err
	}

	return gen.GetResults(query)
}

func (gen *Generator) GetResults(query string) (*types.Unreleased, []*types.Version, error) {
	back, err := gen.workdir()
	if err != nil {
		return nil, nil, err
	}
	defer back()

	tags, first, err := gen.getTags(query)
	if err != nil {
		return nil, nil, err
	}

	unreleased, err := gen.readUnreleased(tags, gen.processor)
	if err != nil {
		return nil, nil, err
	}

	versions, err := gen.readVersions(tags, first, gen.processor)
	if err != nil {
		return nil, nil, err
	}

	if len(versions) == 0 {
		return nil, nil, errors.Errorf("commits corresponding to %q was not found", query)
	}

	return unreleased, versions, nil
}

func (gen *Generator) readVersions(tags []*types.Tag, first string, processor gitlib.Processor) ([]*types.Version, error) {
	next := gen.config.Options.NextTag
	versions := []*types.Version{}

	for i, tag := range tags {
		var (
			isNext = next == tag.Name
			rev    string
		)

		if isNext {
			if tag.Previous != nil {
				rev = tag.Previous.Name + "..HEAD"
			} else {
				rev = "HEAD"
			}
		} else {
			if i+1 < len(tags) {
				rev = tags[i+1].Name + ".." + tag.Name
			} else {
				if first != "" {
					rev = first + ".." + tag.Name
				} else {
					rev = tag.Name
				}
			}
		}

		commits, err := gen.commitParser.Parse(rev, processor)
		if err != nil {
			return nil, err
		}

		commitGroups, mergeCommits, revertCommits, noteGroups := gen.commitExtractor.Extract(commits)

		versions = append(versions, &types.Version{
			Tag:           tag,
			CommitGroups:  commitGroups,
			Commits:       commits,
			MergeCommits:  mergeCommits,
			RevertCommits: revertCommits,
			NoteGroups:    noteGroups,
		})

		// Instead of `getTags()`, assign the date to the tag
		if isNext && len(commits) != 0 {
			tag.Date = commits[0].Author.Date
		}
	}

	return versions, nil
}

func (gen *Generator) readUnreleased(tags []*types.Tag, processor gitlib.Processor) (*types.Unreleased, error) {
	if gen.config.Options.NextTag != "" {
		return &types.Unreleased{}, nil
	}

	rev := "HEAD"

	if len(tags) > 0 {
		rev = tags[0].Name + "..HEAD"
	}

	commits, err := gen.commitParser.Parse(rev, processor)
	if err != nil {
		return nil, err
	}

	commitGroups, mergeCommits, revertCommits, noteGroups := gen.commitExtractor.Extract(commits)

	unreleased := &types.Unreleased{
		CommitGroups:  commitGroups,
		Commits:       commits,
		MergeCommits:  mergeCommits,
		RevertCommits: revertCommits,
		NoteGroups:    noteGroups,
	}

	return unreleased, nil
}

func (gen *Generator) getTags(query string) ([]*types.Tag, string, error) {
	tags, err := gen.tagReader.ReadAll()
	if err != nil {
		return nil, "", errors.Wrap(err, "read all tags")
	}

	next := gen.config.Options.NextTag
	if next != "" {
		for _, tag := range tags {
			if next == tag.Name {
				return nil, "", errors.Errorf("\"%s\" tag already exists", next)
			}
		}

		var previous *types.RelateTag
		if len(tags) > 0 {
			previous = &types.RelateTag{
				Name:    tags[0].Name,
				Subject: tags[0].Subject,
				Date:    tags[0].Date,
			}
		}

		// Assign the date with `readVersions()`
		tags = append([]*types.Tag{
			{
				Name:     next,
				Subject:  next,
				Previous: previous,
			},
		}, tags...)
	}

	if len(tags) == 0 {
		return nil, "", errors.Errorf("git-tag does not exist")
	}

	first := ""
	if query != "" {
		tags, first, err = gen.tagSelector.Select(tags, query)
		if err != nil {
			return nil, "", err
		}
	}

	return tags, first, nil
}

func (gen *Generator) render(w io.Writer, unreleased *types.Unreleased, versions []*types.Version) error {
	if _, err := os.Stat(gen.config.Template); err != nil {
		return err
	}

	fname := filepath.Base(gen.config.Template)

	t := template.Must(template.New(fname).Funcs(TemplateFuncMap).ParseFiles(gen.config.Template))

	return t.Execute(w, &types.RenderData{
		Info:       gen.config.Info,
		Unreleased: unreleased,
		Versions:   versions,
	})
}
