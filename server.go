package main

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"net"
	"net/http"
)

//go:embed index.html
var templateFS embed.FS

var pageTemplate = template.Must(
	template.ParseFS(templateFS, "index.html"),
)

type Server struct {
	matcher *Matcher
}

func NewServer(config *Config) (*Server, error) {
	matcher, err := NewMatcher(config)
	if err != nil {
		return nil, fmt.Errorf("create matcher: %w", err)
	}

	return &Server{
		matcher: matcher,
	}, nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	importPath := requestImportPath(r)

	rule, ok := s.matcher.Match(importPath)
	if !ok {
		http.NotFound(w, r)
		return
	}

	data := pageData{
		ImportPath: importPath,
		Import:     formatImport(&rule.Import),
		Source:     formatSource(rule.Source),
	}

	var body bytes.Buffer

	if err := pageTemplate.Execute(&body, data); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	_, _ = w.Write(body.Bytes())
}

func requestImportPath(r *http.Request) string {
	host := r.Host

	if hostName, _, err := net.SplitHostPort(host); err == nil {
		host = hostName
	}

	return host + r.URL.Path
}

type pageData struct {
	ImportPath string
	Import     string
	Source     string
}

func formatImport(importRule *ImportRule) string {
	if importRule.Subdirectory == nil {
		return fmt.Sprintf("%s %s", importRule.VCS, importRule.Repo)
	}

	return fmt.Sprintf(
		"%s %s %s",
		importRule.VCS,
		importRule.Repo,
		*importRule.Subdirectory,
	)
}

func formatSource(source *Source) string {
	if source == nil {
		return ""
	}

	return fmt.Sprintf("%s %s %s", source.Web, source.File, source.Line)
}
