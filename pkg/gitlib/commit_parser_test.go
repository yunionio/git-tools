package gitlib

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/yunionio/git-tools/pkg/types"
)

func TestCommitParserParse(t *testing.T) {
	assert := assert.New(t)
	assert.True(true)

	mock := &mockClient{
		ReturnExec: func(subcmd string, args ...string) (string, error) {
			if subcmd != "log" {
				return "", errors.New("")
			}

			bytes, _ := ioutil.ReadFile(filepath.Join("testdata", "gitlog.txt"))

			return string(bytes), nil
		},
	}

	parser := NewCommitParser(mock, &types.ChangelogConfig{
		Options: &types.ChangelogConfigOptions{
			CommitFilters: map[string][]string{
				"Type": {
					"feat",
					"fix",
					"perf",
					"refactor",
				},
			},
			HeaderPattern: "^(\\w*)(?:\\(([\\w\\$\\.\\-\\*\\s]*)\\))?\\:\\s(.*)$",
			HeaderPatternMaps: []string{
				"Type",
				"Scope",
				"Subject",
			},
			IssuePrefix: []string{
				"#",
				"gh-",
			},
			RefActions: []string{
				"close",
				"closes",
				"closed",
				"fix",
				"fixes",
				"fixed",
				"resolve",
				"resolves",
				"resolved",
			},
			MergePattern: "^Merge pull request #(\\d+) from (.*)$",
			MergePatternMaps: []string{
				"Ref",
				"Source",
			},
			RevertPattern: "^Revert \"([\\s\\S]*)\"$",
			RevertPatternMaps: []string{
				"Header",
			},
			NoteKeywords: []string{
				"BREAKING CHANGE",
			},
		},
	})

	commits, err := parser.Parse("HEAD", nil)
	assert.Nil(err)
	assert.Equal([]*types.Commit{
		{
			Hash: &types.CommitHash{
				Long:  "65cf1add9735dcc4810dda3312b0792236c97c4e",
				Short: "65cf1add",
			},
			Author: &types.CommitAuthor{
				Name:  "tsuyoshi wada",
				Email: "mail@example.com",
				Date:  time.Unix(int64(1514808000), 0),
			},
			Committer: &types.CommitCommitter{
				Name:  "tsuyoshi wada",
				Email: "mail@example.com",
				Date:  time.Unix(int64(1514808000), 0),
			},
			Merge:  nil,
			Revert: nil,
			Refs: []*types.CommitRef{
				{
					Action: "",
					Ref:    "123",
					Source: "",
				},
			},
			Notes:    []*types.CommitNote{},
			Mentions: []string{},
			Header:   "feat(*): Add new feature #123",
			Type:     "feat",
			Scope:    "*",
			Subject:  "Add new feature #123",
			Body:     "",
		},
		{
			Hash: &types.CommitHash{
				Long:  "14ef0b6d386c5432af9292eab3c8314fa3001bc7",
				Short: "14ef0b6d",
			},
			Author: &types.CommitAuthor{
				Name:  "tsuyoshi wada",
				Email: "mail@example.com",
				Date:  time.Unix(int64(1515153600), 0),
			},
			Committer: &types.CommitCommitter{
				Name:  "tsuyoshi wada",
				Email: "mail@example.com",
				Date:  time.Unix(int64(1515153600), 0),
			},
			Merge: &types.CommitMerge{
				Ref:    "3",
				Source: "username/branchname",
			},
			Revert: nil,
			Refs: []*types.CommitRef{
				{
					Action: "",
					Ref:    "3",
					Source: "",
				},
				{
					Action: "Fixes",
					Ref:    "3",
					Source: "",
				},
				{
					Action: "Closes",
					Ref:    "1",
					Source: "",
				},
			},
			Notes: []*types.CommitNote{
				{
					Title: "BREAKING CHANGE",
					Body:  "This is breaking point message.",
				},
			},
			Mentions: []string{},
			Header:   "Merge pull request #3 from username/branchname",
			Type:     "",
			Scope:    "",
			Subject:  "",
			Body: `This is body message.

Fixes #3

Closes #1

BREAKING CHANGE: This is breaking point message.`,
		},
		{
			Hash: &types.CommitHash{
				Long:  "809a8280ffd0dadb0f4e7ba9fc835e63c37d6af6",
				Short: "809a8280",
			},
			Author: &types.CommitAuthor{
				Name:  "tsuyoshi wada",
				Email: "mail@example.com",
				Date:  time.Unix(int64(1517486400), 0),
			},
			Committer: &types.CommitCommitter{
				Name:  "tsuyoshi wada",
				Email: "mail@example.com",
				Date:  time.Unix(int64(1517486400), 0),
			},
			Merge:  nil,
			Revert: nil,
			Refs:   []*types.CommitRef{},
			Notes:  []*types.CommitNote{},
			Mentions: []string{
				"tsuyoshiwada",
				"hogefuga",
				"FooBarBaz",
			},
			Header:  "fix(controller): Fix cors configure",
			Type:    "fix",
			Scope:   "controller",
			Subject: "Fix cors configure",
			Body: `Has mention body

@tsuyoshiwada
@hogefuga
@FooBarBaz`,
		},
		{
			Hash: &types.CommitHash{
				Long:  "74824d6bd1470b901ec7123d13a76a1b8938d8d0",
				Short: "74824d6b",
			},
			Author: &types.CommitAuthor{
				Name:  "tsuyoshi wada",
				Email: "mail@example.com",
				Date:  time.Unix(int64(1517488587), 0),
			},
			Committer: &types.CommitCommitter{
				Name:  "tsuyoshi wada",
				Email: "mail@example.com",
				Date:  time.Unix(int64(1517488587), 0),
			},
			Merge:  nil,
			Revert: nil,
			Refs: []*types.CommitRef{
				{
					Action: "Fixes",
					Ref:    "123",
					Source: "",
				},
				{
					Action: "Closes",
					Ref:    "456",
					Source: "username/repository",
				},
			},
			Notes: []*types.CommitNote{
				{
					Title: "BREAKING CHANGE",
					Body: fmt.Sprintf(`This is multiline breaking change note.
It is treated as the body of the Note until a mention or reference appears.

We also allow blank lines :)

Example:

%sjavascript
import { Controller } from 'hoge-fuga';

@autobind
class MyController extends Controller {
  constructor() {
    super();
  }
}
%s`, "```", "```"),
				},
			},
			Mentions: []string{},
			Header:   "fix(model): Remove hoge attributes",
			Type:     "fix",
			Scope:    "model",
			Subject:  "Remove hoge attributes",
			Body: fmt.Sprintf(`This mixed body message.

BREAKING CHANGE:
This is multiline breaking change note.
It is treated as the body of the Note until a mention or reference appears.

We also allow blank lines :)

Example:

%sjavascript
import { Controller } from 'hoge-fuga';

@autobind
class MyController extends Controller {
  constructor() {
    super();
  }
}
%s

Fixes #123
Closes username/repository#456`, "```", "```"),
		},
		{
			Hash: &types.CommitHash{
				Long:  "123456789735dcc4810dda3312b0792236c97c4e",
				Short: "12345678",
			},
			Author: &types.CommitAuthor{
				Name:  "tsuyoshi wada",
				Email: "mail@example.com",
				Date:  time.Unix(int64(1517488587), 0),
			},
			Committer: &types.CommitCommitter{
				Name:  "tsuyoshi wada",
				Email: "mail@example.com",
				Date:  time.Unix(int64(1517488587), 0),
			},
			Merge: nil,
			Revert: &types.CommitRevert{
				Header: "fix(core): commit message",
			},
			Refs:     []*types.CommitRef{},
			Notes:    []*types.CommitNote{},
			Mentions: []string{},
			Header:   "Revert \"fix(core): commit message\"",
			Type:     "",
			Scope:    "",
			Subject:  "",
			Body:     "This reverts commit f755db78dcdf461dc42e709b3ab728ceba353d1d.",
		},
	}, commits)
}
