---
title: "v{{.TagName}}"
weight: -{{.Weight}}
---

发布时间 {{ datetime "2006-01-02 15:04:05" .Date }}

{{ range .Repos -}}
{{ if isCommitsNotEmpty .Commits -}}

-----

## {{ .Repo.Name }}

仓库地址: {{ .Repo.URL }}

{{ len .Commits }} commits to {{ tagNameRef .Repo.Name .Tag }} since this release.

{{ range .CommitGroups -}}
### {{ .Title }} ({{len .Commits}})
{{ range .Commits -}}
- {{ commitSummary . }}
{{ end }}
{{ end -}}

{{- if .NoteGroups -}}
{{ range .NoteGroups -}}
### {{ .Title }}
{{ range .Notes }}
{{ .Body }}
{{ end }}
{{ end -}}
{{ end -}}

{{ if .Tag.Previous -}}
{{ tagRef .Tag .Repo.Name .Repo.URL }}
{{ end -}}
{{ end -}}
{{ end -}}
