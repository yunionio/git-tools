package gitlib

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/blang/semver/v4"
	gitcmd "github.com/tsuyoshiwada/go-gitcmd"

	"yunion.io/x/log"

	"github.com/yunionio/git-tools/pkg/types"
)

type TagReader interface {
	ReadAll() ([]*types.Tag, error)
}

type tagReader struct {
	client    gitcmd.Client
	format    string
	separator string
	reFilter  *regexp.Regexp
	useSemVer bool
}

func NewTagReader(client gitcmd.Client, filterPattern string) *tagReader {
	return &tagReader{
		client:    client,
		separator: separator,
		reFilter:  regexp.MustCompile(filterPattern),
		useSemVer: false,
	}
}

func NewSemVerTagReader(client gitcmd.Client) *tagReader {
	return &tagReader{
		client:    client,
		separator: separator,
		reFilter:  regexp.MustCompile("^v"),
		useSemVer: true,
	}
}

func (r *tagReader) ReadAll() ([]*types.Tag, error) {
	out, err := r.client.Exec(
		"for-each-ref",
		"--format",
		"%(refname)"+r.separator+"%(subject)"+r.separator+"%(taggerdate)"+r.separator+"%(authordate)",
		"refs/tags",
	)

	tags := []*types.Tag{}

	if err != nil {
		return tags, fmt.Errorf("failed to get git-tag: %s", err.Error())
	}

	lines := strings.Split(out, "\n")

	for _, line := range lines {
		tokens := strings.Split(line, r.separator)

		if len(tokens) != 4 {
			continue
		}

		name := r.parseRefname(tokens[0])
		subject := r.parseSubject(tokens[1])
		date, err := r.parseDate(tokens[2])
		if err != nil {
			t, err2 := r.parseDate(tokens[3])
			if err2 != nil {
				return nil, err2
			}
			date = t
		}

		if r.reFilter != nil {
			if !r.reFilter.MatchString(name) {
				continue
			}
		}

		var ver *semver.Version
		semverReg := regexp.MustCompile(`^v[\d]+\.[\d]+\.[\d]+$`)
		if r.useSemVer {
			vName := strings.TrimPrefix(name, "v")
			verObj, err := semver.Make(vName)
			if err != nil || !semverReg.MatchString(name) {
				log.Warningf("tag %s is not semver, skip it", name)
				continue
			}
			ver = &verObj
		}

		tags = append(tags, &types.Tag{
			Name:    name,
			Subject: subject,
			Date:    date,
			Version: ver,
		})
	}

	r.sortTags(tags, r.useSemVer)
	r.assignPreviousAndNextTag(tags)

	return tags, nil
}

func (*tagReader) parseRefname(input string) string {
	return strings.Replace(input, "refs/tags/", "", 1)
}

func (*tagReader) parseSubject(input string) string {
	return strings.TrimSpace(input)
}

func (*tagReader) parseDate(input string) (time.Time, error) {
	return time.ParseInLocation("Mon Jan 2 15:04:05 2006 -0700", input, time.UTC)
}

func (*tagReader) assignPreviousAndNextTag(tags []*types.Tag) {
	total := len(tags)

	for i, tag := range tags {
		var (
			next *types.RelateTag
			prev *types.RelateTag
		)

		if i > 0 {
			next = &types.RelateTag{
				Name:    tags[i-1].Name,
				Subject: tags[i-1].Subject,
				Date:    tags[i-1].Date,
			}
		}

		if i+1 < total {
			prev = &types.RelateTag{
				Name:    tags[i+1].Name,
				Subject: tags[i+1].Subject,
				Date:    tags[i+1].Date,
			}
		}

		tag.Next = next
		tag.Previous = prev
	}
}

func (*tagReader) sortTags(tags []*types.Tag, useSemVer bool) {
	sort.Slice(tags, func(i, j int) bool {
		if useSemVer {
			return tags[i].Version.GE(*tags[j].Version)
		}
		return !tags[i].Date.Before(tags[j].Date)
	})
}
