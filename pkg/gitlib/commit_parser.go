package gitlib

// code references from https://github.com/git-chglog/git-chglog
import (
	"regexp"
	"strconv"
	"strings"
	"time"

	gitcmd "github.com/tsuyoshiwada/go-gitcmd"

	"github.com/yunionio/git-tools/pkg/types"
)

var (
	// constants
	separator = "@@__CHGLOG__@@"
	delimiter = "@@__CHGLOG_DELIMITER__@@"

	// fields
	hashField      = "HASH"
	authorField    = "AUTHOR"
	committerField = "COMMITTER"
	subjectField   = "SUBJECT"
	bodyField      = "BODY"

	// formats
	hashFormat      = hashField + ":%H\t%h"
	authorFormat    = authorField + ":%an\t%ae\t%at"
	committerFormat = committerField + ":%cn\t%ce\t%ct"
	subjectFormat   = subjectField + ":%s"
	bodyFormat      = bodyField + ":%b"

	// log
	logFormat = separator + strings.Join([]string{
		hashFormat,
		authorFormat,
		committerFormat,
		subjectFormat,
		bodyFormat,
	}, delimiter)
)

func joinAndQuoteMeta(list []string, sep string) string {
	arr := make([]string, len(list))
	for i, s := range list {
		arr[i] = regexp.QuoteMeta(s)
	}
	return strings.Join(arr, sep)
}

type CommitParser interface {
	Parse(rev string, processor Processor) ([]*types.Commit, error)
}

type commitParser struct {
	client    gitcmd.Client
	config    *types.ChangelogConfig
	reHeader  *regexp.Regexp
	reMerge   *regexp.Regexp
	reRevert  *regexp.Regexp
	reRef     *regexp.Regexp
	reIssue   *regexp.Regexp
	reNotes   *regexp.Regexp
	reMention *regexp.Regexp
}

func NewCommitParser(client gitcmd.Client, config *types.ChangelogConfig) CommitParser {
	opts := config.Options

	joinedRefActions := joinAndQuoteMeta(opts.RefActions, "|")
	joinedIssuePrefix := joinAndQuoteMeta(opts.IssuePrefix, "|")
	joinedNoteKeywords := joinAndQuoteMeta(opts.NoteKeywords, "|")

	return &commitParser{
		client:    client,
		config:    config,
		reHeader:  regexp.MustCompile(opts.HeaderPattern),
		reMerge:   regexp.MustCompile(opts.MergePattern),
		reRevert:  regexp.MustCompile(opts.RevertPattern),
		reRef:     regexp.MustCompile("(?i)(" + joinedRefActions + ")\\s?([\\w/\\.\\-]+)?(?:" + joinedIssuePrefix + ")(\\d+)"),
		reIssue:   regexp.MustCompile("(?:" + joinedIssuePrefix + ")(\\d+)"),
		reNotes:   regexp.MustCompile("^(?i)\\s*(" + joinedNoteKeywords + ")[:\\s]+(.*)"),
		reMention: regexp.MustCompile("@([\\w-]+)"),
	}
}

func (p *commitParser) Parse(rev string, processor Processor) ([]*types.Commit, error) {
	args := []string{}
	if p.config.Options.NoMerges {
		args = append(args, "--no-merges")
	}
	args = append(args, rev, "--no-decorate", "--pretty="+logFormat)
	out, err := p.client.Exec(
		"log",
		args...,
	)

	if err != nil {
		return nil, err
	}

	lines := strings.Split(out, separator)
	lines = lines[1:]
	commits := make([]*types.Commit, len(lines))

	for i, line := range lines {
		commit := p.parseCommit(line)

		if processor != nil {
			commit = processor.ProcessCommit(commit)
			if commit == nil {
				continue
			}
		}

		commits[i] = commit
	}

	return commits, nil
}

func (p *commitParser) parseCommit(input string) *types.Commit {
	commit := &types.Commit{}
	tokens := strings.Split(input, delimiter)

	for _, token := range tokens {
		firstSep := strings.Index(token, ":")
		field := token[0:firstSep]
		value := strings.TrimSpace(token[firstSep+1:])

		switch field {
		case hashField:
			commit.Hash = p.parseHash(value)
		case authorField:
			commit.Author = p.parseAuthor(value)
		case committerField:
			commit.Committer = p.parseCommitter(value)
		case subjectField:
			p.processHeader(commit, value)
		case bodyField:
			p.processBody(commit, value)
		}
	}

	commit.Refs = p.uniqRefs(commit.Refs)
	commit.Mentions = p.uniqMentions(commit.Mentions)

	return commit
}

func (p *commitParser) parseHash(input string) *types.CommitHash {
	arr := strings.Split(input, "\t")

	return &types.CommitHash{
		Long:  arr[0],
		Short: arr[1],
	}
}

func (p *commitParser) parseAuthor(input string) *types.CommitAuthor {
	arr := strings.Split(input, "\t")
	ts, err := strconv.Atoi(arr[2])
	if err != nil {
		ts = 0
	}

	return &types.CommitAuthor{
		Name:  arr[0],
		Email: arr[1],
		Date:  time.Unix(int64(ts), 0),
	}
}

func (p *commitParser) parseCommitter(input string) *types.CommitCommitter {
	author := p.parseAuthor(input)

	return &types.CommitCommitter{
		Name:  author.Name,
		Email: author.Email,
		Date:  author.Date,
	}
}

func (p *commitParser) processHeader(commit *types.Commit, input string) {
	opts := p.config.Options

	// header (raw)
	commit.Header = input

	var res [][]string

	// Type, Scope, Subject etc ...
	res = p.reHeader.FindAllStringSubmatch(input, -1)
	if len(res) > 0 {
		assignDynamicValues(commit, opts.HeaderPatternMaps, res[0][1:])
	}

	// Merge
	res = p.reMerge.FindAllStringSubmatch(input, -1)
	if len(res) > 0 {
		merge := &types.CommitMerge{}
		assignDynamicValues(merge, opts.MergePatternMaps, res[0][1:])
		commit.Merge = merge
	}

	// Revert
	res = p.reRevert.FindAllStringSubmatch(input, -1)
	if len(res) > 0 {
		revert := &types.CommitRevert{}
		assignDynamicValues(revert, opts.RevertPatternMaps, res[0][1:])
		commit.Revert = revert
	}

	// refs & mentions
	commit.Refs = p.parseRefs(input)
	commit.Mentions = p.parseMentions(input)
}

func (p *commitParser) processBody(commit *types.Commit, input string) {
	input = convNewline(input, "\n")

	// body
	commit.Body = input

	// notes & refs & mentions
	commit.Notes = []*types.CommitNote{}
	inNote := false
	fenceDetector := newMdFenceDetector()
	lines := strings.Split(input, "\n")

	for _, line := range lines {
		fenceDetector.Update(line)

		if !fenceDetector.InCodeblock() {
			refs := p.parseRefs(line)
			if len(refs) > 0 {
				inNote = false
				commit.Refs = append(commit.Refs, refs...)
			}

			mentions := p.parseMentions(line)
			if len(mentions) > 0 {
				inNote = false
				commit.Mentions = append(commit.Mentions, mentions...)
			}
		}

		res := p.reNotes.FindAllStringSubmatch(line, -1)

		if len(res) > 0 {
			inNote = true
			for _, r := range res {
				commit.Notes = append(commit.Notes, &types.CommitNote{
					Title: r[1],
					Body:  r[2],
				})
			}
		} else if inNote {
			last := commit.Notes[len(commit.Notes)-1]
			last.Body = last.Body + "\n" + line
		}
	}

	p.trimSpaceInNotes(commit)
}

func (*commitParser) trimSpaceInNotes(commit *types.Commit) {
	for _, note := range commit.Notes {
		note.Body = strings.TrimSpace(note.Body)
	}
}

func (p *commitParser) parseRefs(input string) []*types.CommitRef {
	refs := []*types.CommitRef{}

	// references
	res := p.reRef.FindAllStringSubmatch(input, -1)

	for _, r := range res {
		refs = append(refs, &types.CommitRef{
			Action: r[1],
			Source: r[2],
			Ref:    r[3],
		})
	}

	// issues
	res = p.reIssue.FindAllStringSubmatch(input, -1)
	for _, r := range res {
		duplicate := false
		for _, ref := range refs {
			if ref.Ref == r[1] {
				duplicate = true
			}
		}
		if !duplicate {
			refs = append(refs, &types.CommitRef{
				Action: "",
				Source: "",
				Ref:    r[1],
			})
		}
	}

	return refs
}

func (p *commitParser) parseMentions(input string) []string {
	res := p.reMention.FindAllStringSubmatch(input, -1)
	mentions := make([]string, len(res))

	for i, r := range res {
		mentions[i] = r[1]
	}

	return mentions
}

func (p *commitParser) uniqRefs(refs []*types.CommitRef) []*types.CommitRef {
	arr := []*types.CommitRef{}

	for _, ref := range refs {
		exist := false
		for _, r := range arr {
			if ref.Ref == r.Ref && ref.Action == r.Action && ref.Source == r.Source {
				exist = true
			}
		}
		if !exist {
			arr = append(arr, ref)
		}
	}

	return arr
}

func (p *commitParser) uniqMentions(mentions []string) []string {
	arr := []string{}

	for _, mention := range mentions {
		exist := false
		for _, m := range arr {
			if mention == m {
				exist = true
			}
		}
		if !exist {
			arr = append(arr, mention)
		}
	}

	return arr
}

var (
	fenceTypes = []string{
		"```",
		"~~~",
		"    ",
		"\t",
	}
)

type mdFenceDetector struct {
	fence int
}

func newMdFenceDetector() *mdFenceDetector {
	return &mdFenceDetector{
		fence: -1,
	}
}

func (d *mdFenceDetector) InCodeblock() bool {
	return d.fence > -1
}

func (d *mdFenceDetector) Update(input string) {
	for i, s := range fenceTypes {
		if d.fence < 0 {
			if strings.Index(input, s) == 0 {
				d.fence = i
				break
			}
		} else {
			if strings.Index(input, s) == 0 && i == d.fence {
				d.fence = -1
				break
			}
		}
	}
}
