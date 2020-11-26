package types

import (
	"sort"
	"strings"
	"time"

	"github.com/blang/semver/v4"
)

type GlobalChangeLogConfig struct {
	// Bin is git execution command
	Bin string `json:"bin"`
	// Path for template file
	Template string
	// CacheDir for local repository clone directory
	CacheDir string `json:"cacheDir"`
	// Releases is each release branch want to generate changelog
	Releases []*ReleaseChangeLogConfig `json:"releases"`
	// Options configure generate changelog options
	Options *ChangelogConfigOptions `json:"options"`
	// Output configure output handle options
	Output *GlobalChangelogOutConfig `json:"output"`
}

func (gConf GlobalChangeLogConfig) ToChangelogConfig(rls ReleaseChangeLogConfig, repoIdx int) *ChangelogConfig {
	return rls.ToChangelogConfig(gConf.Bin, gConf.Options, repoIdx)
}

type ReleaseChangeLogConfig struct {
	Branch string        `json:"branch"`
	Repos  []*Repository `json:"repos"`
}

func (rConf ReleaseChangeLogConfig) ToChangelogConfig(bin string, opts *ChangelogConfigOptions, repoIdx int) *ChangelogConfig {
	repo := rConf.Repos[repoIdx]
	return &ChangelogConfig{
		Bin:        bin,
		WorkingDir: repo.WorkingDir,
		Info: &ChangelogConfigInfo{
			RepositoryURL: repo.URL,
		},
		Options: opts,
	}
}

type Repository struct {
	// URL is repo remote url, e.g. `https://github.com/yunionio/onecloud`
	URL string `json:"url"`
	// Working directory
	WorkingDir string `json:"workingDir"`
	// Name is optional, set by url if empty
	Name string `json:"name"`
	// Kind is optional, e.g. `BE` or `FE`
	// Kind string `json:"kind"`
}

type ChangelogConfig struct {
	// Bin is git execution command
	Bin string `json:"bin"`
	// Working directory
	WorkingDir string `json:"workingDir"`
	// Path for template file. If a relative path is specified, it depends on the value of `WorkingDir`.
	Template string

	Info    *ChangelogConfigInfo    `json:"info"`
	Options *ChangelogConfigOptions `json:"options"`
}

type ChangelogConfigInfo struct {
	// Title of CHANGELOG
	Title string `json:"title"`
	// URL of git repository
	RepositoryURL string `json:"repositoryURL"`
}

type ChangelogConfigOptions struct {
	// Treat unreleased commits as specified tags (EXPERIMENTAL)
	NextTag string `json:"nextTag"`
	// Use semantic versioning sort tag
	UseSemVer bool `json:"useSemVer"`
	// Filter tag by regexp
	TagFilterPattern string `json:"tagFilterPattern"`
	// Filter commits in a case insensitive way
	NoCaseSensitive bool `json:"noCaseSensitive"`
	// Filter by using `Commit` properties and values. Filtering is not done by specifying an empty value
	CommitFilters map[string][]string `json:"commitFilters"`
	// Property name to use for sorting `Commit` (e.g. `Scope`)
	CommitSortBy string `json:"commitSortBy"`
	// Property name of `Commit` to be grouped into `CommitGroup` (e.g. `Type`)
	CommitGroupBy string `json:"commitGroupBy"`
	// Property name to use for sorting `CommitGroup` (e.g. `Title`)
	CommitGroupSortBy string `json:"commitGroupSortBy"`
	// Map for `CommitGroup` title conversion
	CommitGroupTitleMaps map[string]string `json:"commitGroupTitleMaps"`
	// A regular expression to use for parsing the commit header
	HeaderPattern string `json:"headerPattern"`
	// A rule for mapping the result of `HeaderPattern` to the property of `Commit`
	HeaderPatternMaps []string `json:"headerPatternMaps"`
	// Prefix used for issues (e.g. `#`, `gh-`)
	IssuePrefix []string `json:"issuePrefix"`
	// Word list of `Ref.Action`
	RefActions []string `json:"refActions"`
	// Fetch logs no merges
	NoMerges bool `json:"noMerges"`
	// A regular expression to use for parsing the merge commit
	MergePattern string `json:"mergePattern"`
	// Similar to `HeaderPatternMaps`
	MergePatternMaps []string `json:"mergePatternMaps"`
	// A regular expression to use for parsing the revert commit
	RevertPattern string `json:"revertPattern"`
	// Similar to `HeaderPatternMaps`
	RevertPatternMaps []string `json:"revertPatternMaps"`
	// Keyword list to find `Note`. A semicolon is a separator, like `<keyword>:` (e.g. `BREAKING CHANGE`)
	NoteKeywords []string `json:"noteKeywords"`
}

type GlobalChangelogOutConfig struct {
	// Dir is output dir
	Dir string `json:"dir"`
}

type Commit struct {
	Repo      string           `json:"repo"`
	Hash      *CommitHash      `json:"hash"`
	Author    *CommitAuthor    `json:"author"`
	Committer *CommitCommitter `json:"committer"`
	// If it is not a merge commit, `nil is assigned`
	Merge *CommitMerge `json:"merge"`
	// if it is not a revert commit, `nil` is assigned
	Revert *CommitRevert `json:"revert"`
	Refs   []*CommitRef  `json:"refs"`
	Notes  []*CommitNote `json:"notes"`
	// Name of the user included in the commit header or body
	Mentions []string `json:"mentions"`
	// (e.g. `feat(core): add new feature`)
	Header string `json:"header"`
	// (e.g. `feat`)
	Type string `json:"type"`
	// (e.g. `core`)
	Scope string `json:"scope"`
	// (e.g. `add new feature`)
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

type CommitHash struct {
	Long  string `json:"long"`
	Short string `json:"short"`
}

type CommitAuthor struct {
	Name  string    `json:"name"`
	Email string    `json:"email"`
	Date  time.Time `json:"date"`
}

type CommitCommitter struct {
	Name  string    `json:"name"`
	Email string    `json:"email"`
	Date  time.Time `json:"date"`
}

// CommitMerge info
type CommitMerge struct {
	Ref    string
	Source string
}

// CommitRevert info
type CommitRevert struct {
	Header string
}

type CommitRef struct {
	// (e.g. `Closes`)
	Action string
	// (e.g. `123`)
	Ref string
	// (e.g. `owner/repository`)
	Source string
}

// CommitNote of commit
type CommitNote struct {
	// (e.g. `BREAKING CHANGE`)
	Title string
	// `Note` content body
	Body string
}

// CommitNoteGroup is a collection of `CommitNote` grouped by titles
type CommitNoteGroup struct {
	Title string        `json:"title"`
	Notes []*CommitNote `json:"notes"`
}

// CommitGroup is a collection of commits grouped according to the `CommitGroupBy` option
type CommitGroup struct {
	// Raw title before conversion (e.g. `build`)
	RawTitle string
	// Conversion by `commitGroupTitleMaps` option, or title converted in title case (e.g. `Build`)
	Title   string
	Commits []*Commit
}

// RelateTag is sibling tag data of `Tag`.
// If you give `Tag`, the reference hierarchy will be deepened.
// This struct is used to minimize the hierarchy of references
type RelateTag struct {
	Name    string
	Subject string
	Date    time.Time
}

// Tag is data of git-tag
type Tag struct {
	Name     string
	Subject  string
	Date     time.Time
	Next     *RelateTag
	Previous *RelateTag
	Version  *semver.Version
}

// Version is a tag-separeted datset to be included in CHANGELOG
type Version struct {
	Tag           *Tag               `json:"tag"`
	CommitGroups  []*CommitGroup     `json:"commitGroups"`
	Commits       []*Commit          `json:"commits"`
	MergeCommits  []*Commit          `json:"mergeCommits"`
	RevertCommits []*Commit          `json:"revertCommits"`
	NoteGroups    []*CommitNoteGroup `json:"noteGroups"`
}

// Unreleased is unreleased commit dataset
type Unreleased struct {
	CommitGroups  []*CommitGroup     `json:"commitGroups"`
	Commits       []*Commit          `json:"commits"`
	MergeCommits  []*Commit          `json:"mergeCommits"`
	RevertCommits []*Commit          `json:"revertCommits"`
	NoteGroups    []*CommitNoteGroup `json:"noteGroups"`
}

// RenderData is the data passed to the template
type RenderData struct {
	Info       *ChangelogConfigInfo
	Unreleased *Unreleased
	Versions   []*Version
}

type GlobalChangeLogResult struct {
	Releases []*ReleaseChangeLogResult `json:"releases"`
}

type ReleaseChangeLogResult struct {
	Branch string                 `json:"branch"`
	Weight int                    `json:"-"`
	Repos  []*RepoChangelogResult `json:"repos"`
}

// RepoChangelogResult contains repo version commits
type RepoChangelogResult struct {
	Repo       *Repository `json:"repo"`
	Versions   []*Version  `json:"versions"`
	Unreleased *Unreleased `json:"unreleased"`
}

type GlobalRenderData struct {
	Releases []*ReleaseRenderData
}

type ReleaseRenderData struct {
	Branch   string
	Weight   int
	Versions []*GlobalVersionRenderData
}

type GlobalVersionRenderData struct {
	TagName string
	Date    time.Time
	Weight  int
	Repos   []*RepoVersionRenderData
}

func (data *GlobalVersionRenderData) Sort() {
	sort.Slice(data.Repos, func(i, j int) bool {
		ri := data.Repos[i]
		rj := data.Repos[j]

		// TODO: support repo weight
		// force sort onecloud repo at first now
		if ri.Repo.Name == "onecloud" {
			return true
		}
		if rj.Repo.Name == "onecloud" {
			return true
		}

		return strings.Compare(ri.Repo.Name, rj.Repo.Name) < 0
	})
}

type RepoVersionRenderData struct {
	Repo *Repository
	*Version
}
