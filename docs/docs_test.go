// Package docs pins the documentation to the code it describes: the flags the
// pages document are the flags offline.go defines, and the pages' relative
// links point at files that exist.
package docs

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
)

var (
	flagDef   = regexp.MustCompile(`flag\.\w+\("([a-z-]+)"`)
	flagUse   = regexp.MustCompile("`(--[a-z-]+)`")
	usageLine = regexp.MustCompile(`(?m)^\./offline (\[.*\]) <program>`)
	mdLink    = regexp.MustCompile(`\]\(([^)#:]+\.md)\)`)
)

func read(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// TestDocumentedFlagsMatchCode - getting-started.md's usage line lists exactly
// the flags offline.go defines, and every flag the page describes is real.
func TestDocumentedFlagsMatchCode(t *testing.T) {
	var defined []string
	for _, m := range flagDef.FindAllStringSubmatch(read(t, "../offline.go"), -1) {
		defined = append(defined, "--"+m[1])
	}
	if len(defined) == 0 {
		t.Fatal("no flag definitions found in offline.go")
	}

	page := read(t, "getting-started.md")
	u := usageLine.FindStringSubmatch(page)
	if u == nil {
		t.Fatal("getting-started.md has no ./offline usage line")
	}
	var usage []string
	for _, f := range strings.Fields(u[1]) {
		usage = append(usage, strings.Trim(f, "[]"))
	}
	slices.Sort(usage)
	slices.Sort(defined)
	if !slices.Equal(usage, defined) {
		t.Errorf("usage line documents %v, offline.go defines %v", usage, defined)
	}

	for _, m := range flagUse.FindAllStringSubmatch(page, -1) {
		if !slices.Contains(defined, m[1]) {
			t.Errorf("getting-started.md documents %s, which offline.go does not define", m[1])
		}
	}
}

// TestRelativeLinksResolve - every relative .md link in the pages names a file
// in docs/, so `mkdocs build --strict` and GitHub rendering both find it.
func TestRelativeLinksResolve(t *testing.T) {
	pages, err := filepath.Glob("*.md")
	if err != nil || len(pages) == 0 {
		t.Fatalf("no pages found: %v", err)
	}
	for _, page := range pages {
		for _, m := range mdLink.FindAllStringSubmatch(read(t, page), -1) {
			if _, err := os.Stat(m[1]); err != nil {
				t.Errorf("%s links to %s: %v", page, m[1], err)
			}
		}
	}
}
