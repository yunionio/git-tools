package gitlib

import (
	"sort"
	"strings"

	"github.com/yunionio/git-tools/pkg/types"
)

type CommitExtractor interface {
	Extract(commits []*types.Commit) ([]*types.CommitGroup, []*types.Commit, []*types.Commit, []*types.CommitNoteGroup)
}

type commitExtractor struct {
	opts *types.ChangelogConfigOptions
}

func NewCommitExtractor(opts *types.ChangelogConfigOptions) *commitExtractor {
	return &commitExtractor{
		opts: opts,
	}
}

func (e *commitExtractor) Extract(commits []*types.Commit) ([]*types.CommitGroup, []*types.Commit, []*types.Commit, []*types.CommitNoteGroup) {
	commitGroups := []*types.CommitGroup{}
	noteGroups := []*types.CommitNoteGroup{}
	mergeCommits := []*types.Commit{}
	revertCommits := []*types.Commit{}

	filteredCommits := commitFilter(commits, e.opts.CommitFilters, e.opts.NoCaseSensitive)

	othersGroup := &types.CommitGroup{
		RawTitle: "Others",
		Title:    "Others",
		Commits:  make([]*types.Commit, 0),
	}

	for _, commit := range commits {
		if commit.Merge != nil {
			mergeCommits = append(mergeCommits, commit)
			continue
		}

		if commit.Revert != nil {
			revertCommits = append(revertCommits, commit)
			continue
		}
	}

	for _, commit := range filteredCommits {
		if commit.Merge == nil && commit.Revert == nil {
			isProcessed := e.processCommitGroups(&commitGroups, commit, e.opts.NoCaseSensitive)
			if !isProcessed {
				othersGroup.Commits = append(othersGroup.Commits, commit)
			}
		}

		e.processNoteGroups(&noteGroups, commit)
	}

	if len(othersGroup.Commits) != 0 {
		commitGroups = append(commitGroups, othersGroup)
	}

	e.sortCommitGroups(commitGroups)
	e.sortNoteGroups(noteGroups)

	return commitGroups, mergeCommits, revertCommits, noteGroups
}

func (e *commitExtractor) processCommitGroups(groups *[]*types.CommitGroup, commit *types.Commit, noCaseSensitive bool) bool {
	var group *types.CommitGroup

	// commit group
	raw, ttl := e.commitGroupTitle(commit)

	for _, g := range *groups {
		rawTitleTmp := g.RawTitle
		if noCaseSensitive {
			rawTitleTmp = strings.ToLower(g.RawTitle)
		}

		rawTmp := raw
		if noCaseSensitive {
			rawTmp = strings.ToLower(raw)
		}
		if rawTitleTmp == rawTmp {
			group = g
		}
	}

	if group != nil {
		group.Commits = append(group.Commits, commit)
	} else if raw != "" {
		*groups = append(*groups, &types.CommitGroup{
			RawTitle: raw,
			Title:    ttl,
			Commits:  []*types.Commit{commit},
		})
	} else {
		return false
	}

	return true
}

func (e *commitExtractor) processNoteGroups(groups *[]*types.CommitNoteGroup, commit *types.Commit) {
	if len(commit.Notes) != 0 {
		for _, note := range commit.Notes {
			e.appendNoteToNoteGroups(groups, note)
		}
	}
}

func (e *commitExtractor) appendNoteToNoteGroups(groups *[]*types.CommitNoteGroup, note *types.CommitNote) {
	exist := false

	for _, g := range *groups {
		if g.Title == note.Title {
			exist = true
			g.Notes = append(g.Notes, note)
		}
	}

	if !exist {
		*groups = append(*groups, &types.CommitNoteGroup{
			Title: note.Title,
			Notes: []*types.CommitNote{note},
		})
	}
}

func (e *commitExtractor) commitGroupTitle(commit *types.Commit) (string, string) {
	var (
		raw string
		ttl string
	)

	if title, ok := dotGet(commit, e.opts.CommitGroupBy); ok {
		if v, ok := title.(string); ok {
			raw = v
			if t, ok := e.opts.CommitGroupTitleMaps[v]; ok {
				ttl = t
			} else {
				ttl = strings.Title(raw)
			}
		}
	}

	return raw, ttl
}

func (e *commitExtractor) sortCommitGroups(groups []*types.CommitGroup) {
	// groups
	sort.Slice(groups, func(i, j int) bool {
		var (
			a, b interface{}
			ok   bool
		)

		a, ok = dotGet(groups[i], e.opts.CommitGroupSortBy)
		if !ok {
			return false
		}

		b, ok = dotGet(groups[j], e.opts.CommitGroupSortBy)
		if !ok {
			return false
		}

		res, err := compare(a, "<", b)
		if err != nil {
			return false
		}
		return res
	})

	// commits
	for _, group := range groups {
		sort.Slice(group.Commits, func(i, j int) bool {
			var (
				a, b interface{}
				ok   bool
			)

			a, ok = dotGet(group.Commits[i], e.opts.CommitSortBy)
			if !ok {
				return false
			}

			b, ok = dotGet(group.Commits[j], e.opts.CommitSortBy)
			if !ok {
				return false
			}

			res, err := compare(a, "<", b)
			if err != nil {
				return false
			}
			return res
		})
	}
}

func (e *commitExtractor) sortNoteGroups(groups []*types.CommitNoteGroup) {
	// groups
	sort.Slice(groups, func(i, j int) bool {
		return strings.ToLower(groups[i].Title) < strings.ToLower(groups[j].Title)
	})

	// notes
	for _, group := range groups {
		sort.Slice(group.Notes, func(i, j int) bool {
			return strings.ToLower(group.Notes[i].Title) < strings.ToLower(group.Notes[j].Title)
		})
	}
}
