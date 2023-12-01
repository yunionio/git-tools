package run

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"yunion.io/x/pkg/errors"

	"io/fs"

	"github.com/yunionio/git-tools/pkg/changelog"
	"github.com/yunionio/git-tools/pkg/types"
	"github.com/yunionio/git-tools/pkg/utils"
)

func handleOutput(data *types.GlobalRenderData, templateFile string, config *types.GlobalChangelogOutConfig) error {
	for _, rls := range data.Releases {
		if err := handleReleaseOutput(rls, templateFile, config); err != nil {
			return err
		}
	}
	return nil
}

func handleReleaseOutput(data *types.ReleaseRenderData, templateFile string, config *types.GlobalChangelogOutConfig) error {
	dir := strings.Replace(data.Branch, "/", "-", -1)
	outDir := path.Join(config.Dir, dir)
	if err := utils.EnsureDir(outDir); err != nil {
		return err
	}

	if err := generateDocusaurusIndxMD(outDir, data); err != nil {
		return errors.Wrap(err, "generate docusaurus index.mdx")
	}

	// if err := generateHugoIndexMD(outDir, data); err != nil {
	// 	return errors.Wrap(err, "generate hugo _index.md")
	// }

	for _, version := range data.Versions {
		if err := handleVersion(version, templateFile, outDir, true); err != nil {
			return errors.Wrapf(err, "handle version %q", version.TagName)
		}
	}

	return nil
}

func generateDocusaurusIndxMD(outDir string, data *types.ReleaseRenderData) error {
	fileName := path.Join(outDir, "index.mdx")

	content := `---
sidebar_position: -%d
---

# %s

%s CHANGELOG 汇总，最近发布版本: %s , 时间: %s
	
import IndexDocCardList from '@site/src/components/IndexDocCardList';

<IndexDocCardList />`

	recentVersion := data.Versions[0]
	recentTag := recentVersion.Repos[0]
	tagName := recentTag.Tag.Name
	// date := recentTag.Tag.Date.Format("2006-01-02 15:04:05")
	date := recentTag.Tag.Date.Format("2006-01-02")
	branch := data.Branch
	content = fmt.Sprintf(content, data.Weight, branch, branch, tagName, date)

	if err := ioutil.WriteFile(fileName, []byte(content), 0644); err != nil {
		return errors.Wrapf(err, "write file %q", fileName)
	}

	return nil
}

func generateHugoIndexMD(outDir string, data *types.ReleaseRenderData) error {
	fileName := path.Join(outDir, "_index.md")

	content := `---
title: "%s"
description: >
  %s CHANGELOG 汇总，最近发布版本: %s , 时间: %s
weight: -%d
---`

	recentVersion := data.Versions[0]
	recentTag := recentVersion.Repos[0]
	tagName := recentTag.Tag.Name
	// date := recentTag.Tag.Date.Format("2006-01-02 15:04:05")
	date := recentTag.Tag.Date.Format("2006-01-02")
	branch := data.Branch
	content = fmt.Sprintf(content, branch, branch, tagName, date, data.Weight)

	if err := ioutil.WriteFile(fileName, []byte(content), 0644); err != nil {
		return errors.Wrapf(err, "write file %q", fileName)
	}

	return nil
}

func handleVersion(version *types.GlobalVersionRenderData, templateFile string, outDir string, isForDocus bool) error {
	if _, err := os.Stat(templateFile); err != nil {
		return errors.Wrapf(err, "stat template file")
	}

	fname := filepath.Base(templateFile)

	t := template.Must(template.New(fname).Funcs(changelog.TemplateFuncMap).ParseFiles(templateFile))

	verStr := strings.ReplaceAll(version.TagName, ".", "-")
	outFile := path.Join(outDir, verStr+".md")
	outF, err := utils.OpenOrCreateFile(outFile)
	if err != nil {
		return errors.Wrapf(err, "open or create file %q", outFile)
	}

	if err := t.Execute(outF, version); err != nil {
		return errors.Wrapf(err, "execute template with version: %#v", version)
	}
	if !isForDocus {
		return nil
	}
	outF.Close()

	// post process for docusaurus
	content, err := ioutil.ReadFile(outFile)
	if err != nil {
		return errors.Wrapf(err, "read file %s for docusaurus", outFile)
	}
	newContent := string(content)
	for k, v := range map[string]string{
		`<`: `\<`,
		`>`: `\>`,
		`{`: `\{`,
		`}`: `\}`,
	} {
		newContent = strings.ReplaceAll(newContent, k, v)
	}
	if err := os.WriteFile(outFile, []byte(newContent), fs.FileMode(0644)); err != nil {
		return errors.Wrapf(err, "write back to %s for docusaurus", outFile)
	}

	return nil
}

/*func handleRepoVersion(data *types.RepoVersionRenderData, templateFile string, outDir string) error {
	if _, err := os.Stat(templateFile); err != nil {
		return errors.Wrapf(err, "stat template file")
	}

	fname := filepath.Base(templateFile)

	t := template.Must(template.New(fname).Funcs(changelog.TemplateFuncMap).ParseFiles(templateFile))

	outFile := path.Join(outDir, data.Repo.Name+".md")
	outF, err := utils.OpenOrCreateFile(outFile)
	if err != nil {
		return errors.Wrapf(err, "open or create file %q", outFile)
	}

	return t.Execute(outF, data)
}*/
