package gitlib

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yunionio/git-tools/pkg/types"
)

func TestCommitExtractor(t *testing.T) {
	assert := assert.New(t)

	extractor := NewCommitExtractor(&types.ChangelogConfigOptions{
		CommitSortBy:      "Scope",
		CommitGroupBy:     "Type",
		CommitGroupSortBy: "Title",
		CommitGroupTitleMaps: map[string]string{
			"bar": "BAR",
		},
	})

	fixtures := []*types.Commit{
		// [0]
		{
			Type:   "foo",
			Scope:  "c",
			Header: "1",
			Notes:  []*types.CommitNote{},
		},
		// [1]
		{
			Type:   "foo",
			Scope:  "b",
			Header: "2",
			Notes: []*types.CommitNote{
				{"note1-title", "note1-body"},
				{"note2-title", "note2-body"},
			},
		},
		// [2]
		{
			Type:   "bar",
			Scope:  "d",
			Header: "3",
			Notes: []*types.CommitNote{
				{"note1-title", "note1-body"},
				{"note3-title", "note3-body"},
			},
		},
		// [3]
		{
			Type:   "foo",
			Scope:  "a",
			Header: "4",
			Notes: []*types.CommitNote{
				{"note4-title", "note4-body"},
			},
		},
		// [4]
		{
			Type:   "",
			Scope:  "",
			Header: "Merge1",
			Notes:  []*types.CommitNote{},
			Merge: &types.CommitMerge{
				Ref:    "123",
				Source: "merges/merge1",
			},
		},
		// [5]
		{
			Type:   "",
			Scope:  "",
			Header: "Revert1",
			Notes:  []*types.CommitNote{},
			Revert: &types.CommitRevert{
				Header: "REVERT1",
			},
		},
	}

	commitGroups, mergeCommits, revertCommits, noteGroups := extractor.Extract(fixtures)

	assert.Equal([]*types.CommitGroup{
		{
			RawTitle: "bar",
			Title:    "BAR",
			Commits: []*types.Commit{
				fixtures[2],
			},
		},
		{
			RawTitle: "foo",
			Title:    "Foo",
			Commits: []*types.Commit{
				fixtures[3],
				fixtures[1],
				fixtures[0],
			},
		},
	}, commitGroups)

	assert.Equal([]*types.Commit{
		fixtures[4],
	}, mergeCommits)

	assert.Equal([]*types.Commit{
		fixtures[5],
	}, revertCommits)

	assert.Equal([]*types.CommitNoteGroup{
		{
			Title: "note1-title",
			Notes: []*types.CommitNote{
				fixtures[1].Notes[0],
				fixtures[2].Notes[0],
			},
		},
		{
			Title: "note2-title",
			Notes: []*types.CommitNote{
				fixtures[1].Notes[1],
			},
		},
		{
			Title: "note3-title",
			Notes: []*types.CommitNote{
				fixtures[2].Notes[1],
			},
		},
		{
			Title: "note4-title",
			Notes: []*types.CommitNote{
				fixtures[3].Notes[0],
			},
		},
	}, noteGroups)
}
