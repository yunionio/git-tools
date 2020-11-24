package gitlib

import (
	"strings"

	"yunion.io/x/pkg/errors"

	"github.com/yunionio/git-tools/pkg/types"
)

const (
	ErrNotFoundTag      = errors.Error("could not find the tag")
	ErrFailedQueryParse = errors.Error("failed to parse the query")
)

type TagSelector interface {
	Select(tags []*types.Tag, query string) ([]*types.Tag, string, error)
}

type tagSelector struct{}

func NewTagSelector() TagSelector {
	return &tagSelector{}
}

func (s *tagSelector) Select(tags []*types.Tag, query string) ([]*types.Tag, string, error) {
	tokens := strings.Split(query, "..")

	switch len(tokens) {
	case 1:
		return s.selectSingleTag(tags, tokens[0])
	case 2:
		old := tokens[0]
		new := tokens[1]
		if old == "" && new == "" {
			return nil, "", nil
		} else if old == "" {
			return s.selectBeforeTags(tags, new)
		} else if new == "" {
			return s.selectAfterTags(tags, old)
		}
		return s.selectRangeTags(tags, tokens[0], tokens[1])
	}

	return nil, "", ErrFailedQueryParse
}

func (s *tagSelector) selectSingleTag(tags []*types.Tag, token string) ([]*types.Tag, string, error) {
	var from string

	for i, tag := range tags {
		if tag.Name == token {
			if i+1 < len(tags) {
				from = tags[i+1].Name
			}
			return []*types.Tag{tag}, from, nil
		}
	}

	return nil, "", nil
}

func (*tagSelector) selectBeforeTags(tags []*types.Tag, token string) ([]*types.Tag, string, error) {
	var (
		res    []*types.Tag
		from   string
		enable bool
	)

	for i, tag := range tags {
		if tag.Name == token {
			enable = true
		}

		if enable {
			res = append(res, tag)
			from = ""
			if i+1 < len(tags) {
				from = tags[i+1].Name
			}
		}
	}

	if len(res) == 0 {
		return res, "", errors.Wrapf(ErrNotFoundTag, "selectBeforeTags token: %s", token)
	}

	return res, from, nil
}

func (*tagSelector) selectAfterTags(tags []*types.Tag, token string) ([]*types.Tag, string, error) {
	var (
		res  []*types.Tag
		from string
	)

	for i, tag := range tags {
		res = append(res, tag)
		from = ""
		if i+1 < len(tags) {
			from = tags[i+1].Name
		}

		if tag.Name == token {
			break
		}
	}

	if len(res) == 0 {
		return res, "", errors.Wrapf(ErrNotFoundTag, "selectAfterTags token: %s", token)
	}

	return res, from, nil
}

func (s *tagSelector) selectRangeTags(tags []*types.Tag, old string, new string) ([]*types.Tag, string, error) {
	var (
		res    []*types.Tag
		from   string
		enable bool
	)

	for i, tag := range tags {
		if tag.Name == new {
			enable = true
		}

		if enable {
			from = ""
			if i+1 < len(tags) {
				from = tags[i+1].Name
			}
			res = append(res, tag)
		}

		if tag.Name == old {
			enable = false
		}
	}

	if len(res) == 0 {
		return res, "", errors.Wrapf(ErrNotFoundTag, "old: %q, new: %q", old, new)
	}

	return res, from, nil
}
