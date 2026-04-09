package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDeriveSlug(t *testing.T) {
	t.Parallel()

	tests := []struct {
		filename string
		want     string
	}{
		{filename: "2026-04-09-hello-world.md", want: "hello-world"},
		{filename: "2026-04-09T110000-hello-world.md", want: "hello-world"},
		{filename: "hello-world.md", want: "hello-world"},
	}

	for _, test := range tests {
		t.Run(test.filename, func(t *testing.T) {
			t.Parallel()
			if got := deriveSlug(test.filename); got != test.want {
				t.Fatalf("deriveSlug(%q) = %q, want %q", test.filename, got, test.want)
			}
		})
	}
}

func TestParsePostHonorsFrontMatterSlug(t *testing.T) {
	root := t.TempDir()
	writeTestTemplates(t, root)
	writeFile(t, filepath.Join(root, "posts", "2026-04-09-custom.md"), `---
title: "Hello"
date: "2026-04-09T110000"
tags: []
draft: false
slug: "from-front-matter"
---

Body text.
`)

	builder := newTestBuilder(t, root)
	posts, err := builder.loadPosts()
	if err != nil {
		t.Fatalf("loadPosts() error = %v", err)
	}
	if len(posts) != 1 {
		t.Fatalf("loadPosts() len = %d, want 1", len(posts))
	}
	if posts[0].Slug != "from-front-matter" {
		t.Fatalf("post slug = %q, want %q", posts[0].Slug, "from-front-matter")
	}
}

func TestLoadPostsValidationErrors(t *testing.T) {
	tests := []struct {
		name         string
		postContents []string
		wantErr      string
		setup        func(t *testing.T, root string)
	}{
		{
			name: "missing title",
			postContents: []string{`---
date: "2026-04-09T110000"
tags: []
draft: false
---

Body text.
`},
			wantErr: "title: missing required field",
		},
		{
			name: "empty title",
			postContents: []string{`---
title: ""
date: "2026-04-09T110000"
tags: []
draft: false
---

Body text.
`},
			wantErr: "title: must not be empty",
		},
		{
			name: "bad date",
			postContents: []string{`---
title: "Hello"
date: "2026-04-09T11:00:00"
tags: []
draft: false
---

Body text.
`},
			wantErr: "date: must use YYYY-MM-DDTHHMMSS",
		},
		{
			name: "missing draft",
			postContents: []string{`---
title: "Hello"
date: "2026-04-09T110000"
tags: []
---

Body text.
`},
			wantErr: "draft: missing required field",
		},
		{
			name: "non string tag",
			postContents: []string{`---
title: "Hello"
date: "2026-04-09T110000"
tags: [ok, 7]
draft: false
---

Body text.
`},
			wantErr: "tags: must be a list of strings",
		},
		{
			name: "missing local image",
			postContents: []string{`---
title: "Hello"
date: "2026-04-09T110000"
tags: []
draft: false
---

![Missing](missing.png)
`},
			wantErr: `image: "missing.png" is not a readable local image`,
		},
		{
			name: "duplicate slug",
			postContents: []string{
				`---
title: "First"
date: "2026-04-09T110000"
tags: []
draft: false
slug: "same-slug"
---

Body text.
`,
				`---
title: "Second"
date: "2026-04-10T110000"
tags: []
draft: false
slug: "same-slug"
---

Body text.
`,
			},
			wantErr: `duplicate slug "same-slug"`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			writeTestTemplates(t, root)
			if test.setup != nil {
				test.setup(t, root)
			}

			for i, contents := range test.postContents {
				writeFile(t, filepath.Join(root, "posts", fmt.Sprintf("2026-04-%02d-post-%d.md", i+1, i+1)), contents)
			}

			builder := newTestBuilder(t, root)
			posts, err := builder.loadPosts()
			if err == nil && strings.Contains(test.wantErr, "duplicate slug") {
				err = validatePublishedSlugs(posts)
			}
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", test.wantErr)
			}
			if !strings.Contains(err.Error(), test.wantErr) {
				t.Fatalf("error = %q, want substring %q", err.Error(), test.wantErr)
			}
		})
	}
}

func TestValidPostPassesValidation(t *testing.T) {
	root := t.TempDir()
	writeTestTemplates(t, root)
	writePNGImage(t, filepath.Join(root, "posts", "local.png"), 640, 360)
	writeFile(t, filepath.Join(root, "posts", "2026-04-09-valid.md"), `---
title: "Valid"
date: "2026-04-09T110000"
tags: [blog]
draft: false
---

![Local](local.png)
`)

	builder := newTestBuilder(t, root)
	posts, err := builder.loadPosts()
	if err != nil {
		t.Fatalf("loadPosts() error = %v", err)
	}
	if err := validatePublishedSlugs(posts); err != nil {
		t.Fatalf("validatePublishedSlugs() error = %v", err)
	}
}

func TestPaginatePostsBoundaries(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		count     int
		wantPages int
		wantSizes []int
	}{
		{name: "zero", count: 0, wantPages: 1, wantSizes: []int{0}},
		{name: "ten", count: 10, wantPages: 1, wantSizes: []int{10}},
		{name: "eleven", count: 11, wantPages: 2, wantSizes: []int{10, 1}},
		{name: "twenty five", count: 25, wantPages: 3, wantSizes: []int{10, 10, 5}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			posts := make([]Post, 0, test.count)
			for i := 0; i < test.count; i++ {
				posts = append(posts, Post{Title: fmt.Sprintf("Post %d", i+1)})
			}

			pages := paginatePosts(posts, 10)
			if len(pages) != test.wantPages {
				t.Fatalf("len(pages) = %d, want %d", len(pages), test.wantPages)
			}
			for i, wantSize := range test.wantSizes {
				if got := len(pages[i]); got != wantSize {
					t.Fatalf("len(pages[%d]) = %d, want %d", i, got, wantSize)
				}
			}
		})
	}
}

func TestParseDateValueSupportsOnlyCompactFormat(t *testing.T) {
	t.Parallel()

	want := time.Date(2026, 4, 9, 11, 0, 0, 0, time.UTC)
	got, err := parseDateValue("2026-04-09T110000")
	if err != nil {
		t.Fatalf("parseDateValue(compact) error = %v", err)
	}
	if !got.Equal(want) {
		t.Fatalf("parseDateValue(compact) = %v, want %v", got, want)
	}

	for _, value := range []interface{}{
		"2026-04-09T11:00:00",
		"2026-04-09",
		time.Date(2026, 4, 9, 11, 0, 0, 0, time.UTC),
	} {
		if _, err := parseDateValue(value); err == nil || !strings.Contains(err.Error(), "YYYY-MM-DDTHHMMSS") {
			t.Fatalf("parseDateValue(%v) error = %v, want YYYY-MM-DDTHHMMSS validation", value, err)
		}
	}
}

func TestRenderMarkdownCodeBlocksAndChromaCSS(t *testing.T) {
	root := t.TempDir()
	writeTestTemplates(t, root)
	builder := newTestBuilder(t, root)

	rendered, err := builder.renderMarkdown("```go\nfmt.Println(\"hi\")\n```\n\n```\nplain\n```\n")
	if err != nil {
		t.Fatalf("renderMarkdown() error = %v", err)
	}
	rendered = wrapPlainCodeBlocks(rendered)

	if !strings.Contains(rendered, `<span class=`) {
		t.Fatalf("rendered output missing Chroma span: %s", rendered)
	}
	if !strings.Contains(rendered, `<div class="plain-code"><pre><code>plain`) {
		t.Fatalf("rendered output missing plain code block wrapper: %s", rendered)
	}

	css, err := generateChromaCSS()
	if err != nil {
		t.Fatalf("generateChromaCSS() error = %v", err)
	}
	if !strings.Contains(css, ".chroma") {
		t.Fatalf("generated CSS missing .chroma selector")
	}
	if !strings.Contains(css, ".chroma .k") {
		t.Fatalf("generated CSS missing keyword selector")
	}
}

func TestRewriteImagesGeneratesResponsivePlans(t *testing.T) {
	root := t.TempDir()
	writeTestTemplates(t, root)
	writePNGImage(t, filepath.Join(root, "posts", "local.png"), 1600, 900)
	builder := newTestBuilder(t, root)

	post := Post{
		Slug:       "hello-world",
		SourcePath: filepath.ToSlash(filepath.Join("posts", "2026-04-09-hello-world.md")),
	}

	rendered, err := builder.renderMarkdown("![Local](local.png)\n\n![Remote](https://example.com/image.png)\n")
	if err != nil {
		t.Fatalf("renderMarkdown() error = %v", err)
	}

	rewritten, plans, err := builder.rewriteImages(post, rendered)
	if err != nil {
		t.Fatalf("rewriteImages() error = %v", err)
	}

	if !strings.Contains(rewritten, `srcset="../assets/hello-world/local-480w.png 480w`) {
		t.Fatalf("rewritten output missing local srcset: %s", rewritten)
	}
	if !strings.Contains(rewritten, `sizes="(min-width: 58rem) 54rem, calc(100vw - 2rem)"`) {
		t.Fatalf("rewritten output missing sizes attribute: %s", rewritten)
	}
	if strings.Count(rewritten, `loading="lazy"`) != 2 {
		t.Fatalf("rewritten output missing lazy loading on both images: %s", rewritten)
	}
	if !strings.Contains(rewritten, `src="https://example.com/image.png" loading="lazy"`) {
		t.Fatalf("remote image should remain unchanged apart from lazy loading: %s", rewritten)
	}

	plan, ok := plans["local.png"]
	if !ok {
		t.Fatalf("expected responsive plan for local image")
	}
	if err := emitResponsiveImagePlan(root, plan); err != nil {
		t.Fatalf("emitResponsiveImagePlan() error = %v", err)
	}

	for _, width := range []int{480, 768, 1200, 1600} {
		outputPath := filepath.Join(root, "assets", "hello-world", fmt.Sprintf("local-%dw.png", width))
		if _, err := os.Stat(outputPath); err != nil {
			t.Fatalf("expected generated asset %s: %v", outputPath, err)
		}
	}
}

func TestCleanupGeneratedOutput(t *testing.T) {
	root := t.TempDir()

	writeFile(t, filepath.Join(root, "index.html"), "<!-- "+generatedMarker+" -->")
	writeFile(t, filepath.Join(root, "page", "2", "index.html"), "older page")
	writeFile(t, filepath.Join(root, "assets", "old-post", "hero-480w.png"), "asset")
	writeFile(t, filepath.Join(root, "old-post", "index.html"), "<!-- "+generatedMarker+" -->")
	writeFile(t, filepath.Join(root, "keep-post", "index.html"), "<!doctype html><p>keep</p>")
	writeFile(t, filepath.Join(root, "mathsite", "index.html"), "<!-- "+generatedMarker+" -->")

	if err := cleanupGeneratedOutput(root); err != nil {
		t.Fatalf("cleanupGeneratedOutput() error = %v", err)
	}

	assertMissing(t, filepath.Join(root, "index.html"))
	assertMissing(t, filepath.Join(root, "page"))
	assertMissing(t, filepath.Join(root, "assets"))
	assertMissing(t, filepath.Join(root, "old-post"))

	if _, err := os.Stat(filepath.Join(root, "keep-post")); err != nil {
		t.Fatalf("keep-post should remain: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "mathsite")); err != nil {
		t.Fatalf("protected directory should remain: %v", err)
	}
}

func TestBuildSiteWritesPaginatedOutput(t *testing.T) {
	root := t.TempDir()
	writeTestTemplates(t, root)
	writeFile(t, filepath.Join(root, "posts", "2026-04-12-post-12.md"), `---
title: "Newest"
date: "2026-04-12T110000"
tags: []
draft: false
---

Newest body.
`)
	writeFile(t, filepath.Join(root, "posts", "2026-04-11-post-11.md"), `---
title: "Second newest"
date: "2026-04-11T110000"
tags: []
draft: false
---

Second newest body.
`)
	for i := 1; i <= 9; i++ {
		writeFile(t, filepath.Join(root, "posts", fmt.Sprintf("2026-04-%02d-post-%02d.md", i, i)), fmt.Sprintf(`---
title: "Post %02d"
date: "2026-04-%02dT110000"
tags: []
draft: false
---

Body for post %02d.
`, i, i, i))
	}
	writeFile(t, filepath.Join(root, "posts", "2026-04-03-draft.md"), `---
title: "Draft"
date: "2026-04-03T110000"
tags: []
draft: true
---

Draft body.
`)

	var stderr bytes.Buffer
	if err := buildSite(root, 10, &stderr); err != nil {
		t.Fatalf("buildSite() error = %v\nstderr:\n%s", err, stderr.String())
	}

	indexHTML := readFile(t, filepath.Join(root, "index.html"))
	pageTwoHTML := readFile(t, filepath.Join(root, "page", "2", "index.html"))
	postHTML := readFile(t, filepath.Join(root, "post-12", "index.html"))

	if !strings.Contains(indexHTML, "Newest") || !strings.Contains(indexHTML, "Second newest") {
		t.Fatalf("index.html missing newest posts:\n%s", indexHTML)
	}
	if strings.Contains(indexHTML, "Draft") || strings.Contains(pageTwoHTML, "Draft") {
		t.Fatalf("draft post should not be rendered")
	}
	if strings.Index(indexHTML, "Newest") > strings.Index(indexHTML, "Second newest") {
		t.Fatalf("posts are not sorted newest-first:\n%s", indexHTML)
	}
	if !strings.Contains(indexHTML, "older=/page/2/") {
		t.Fatalf("index.html missing older posts link:\n%s", indexHTML)
	}
	if !strings.Contains(pageTwoHTML, "newer=/") {
		t.Fatalf("page 2 missing newer posts link:\n%s", pageTwoHTML)
	}
	if !strings.Contains(pageTwoHTML, "Post 01") {
		t.Fatalf("page 2 missing oldest post:\n%s", pageTwoHTML)
	}
	if !strings.Contains(postHTML, "Newest body.") {
		t.Fatalf("post page missing rendered content:\n%s", postHTML)
	}
}

func newTestBuilder(t *testing.T, root string) *siteBuilder {
	t.Helper()

	builder, err := newSiteBuilder(root, 10, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("newSiteBuilder() error = %v", err)
	}
	return builder
}

func writeTestTemplates(t *testing.T, root string) {
	t.Helper()

	writeFile(t, filepath.Join(root, "templates", "index.html.tmpl"), `{{ .Marker }}
{{ if .Posts }}{{ range .Posts }}{{ .Title }}|{{ .Excerpt }}
{{ end }}{{ else }}No posts yet.
{{ end }}{{ if .NewerPageURL }}newer={{ .NewerPageURL }}{{ end }}{{ if .OlderPageURL }}older={{ .OlderPageURL }}{{ end }}
`)
	writeFile(t, filepath.Join(root, "templates", "post.html.tmpl"), `{{ .Marker }}
{{ .Title }}
{{ .DisplayDate }}
{{ .Content }}
<style>{{ .ChromaCSS }}</style>
`)
	writeFile(t, filepath.Join(root, "posts", ".keep"), "")
}

func writeFile(t *testing.T, path string, contents string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", path, err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	return string(data)
}

func writePNGImage(t *testing.T, path string, width int, height int) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", path, err)
	}

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	fill := color.RGBA{R: 0x55, G: 0x99, B: 0xff, A: 0xff}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, fill)
		}
	}

	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("Create(%q) error = %v", path, err)
	}
	defer file.Close()

	if err := png.Encode(file, img); err != nil {
		t.Fatalf("png.Encode(%q) error = %v", path, err)
	}
}

func assertMissing(t *testing.T, path string) {
	t.Helper()

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected %s to be removed, stat err = %v", path, err)
	}
}
