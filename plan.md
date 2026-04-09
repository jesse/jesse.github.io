# Blog Implementation Plan

## Scope

Implements the blog described in `research.md`. Key constraints:

- Prose width ~`44rem`; code blocks may widen beyond that column
- Blockquotes styled as callouts, not browser defaults
- Syntax highlighting via Chroma (CSS classes mode, dark theme)
- Local images resized into responsive variants; remote images left alone
- `make new` scaffolds a post and prints the path
- Slug derivation handles `YYYY-MM-DD-` and `YYYY-MM-DDTHHMMSS-` prefixes
- Date parsing accepts only `YYYY-MM-DDTHHMMSS`

## Phase 1: Project scaffolding

- `go mod init`, then `go get` goldmark, goldmark-meta, goldmark-highlighting/v2, misspell, `x/image/draw`
- Create directories: `cmd/build/`, `posts/`, `templates/`
- Create skeleton Makefile with targets: `new`, `build`, `publish`, `deploy`, `clean`

## Phase 2: HTML templates

### `templates/index.html.tmpl`

- Dark-theme `<style>`: `background: #1a1a1a`, `color: #e8e8e8`, `a { color: #5599ff }`, Georgia serif, `line-height: 1.7`, `max-width: 44rem` centered
- `<header>` with blog title + "Projects" links (Math Adventures, Waveform)
- Post listing loop: date, linked title, excerpt
- Pagination: "Newer posts" / "Older posts" links
- `<footer>`

### `templates/post.html.tmpl`

- Same base CSS as index
- `{{ .ChromaCSS }}` in `<style>` — build generates Chroma token rules once, passes as template var
- Blockquotes: left accent border, subtle background tint, softer text, generous spacing
- Inline code: subtle background, monospace, no highlighting
- Fenced code: `.highlight > pre` with darker surface, monospace stack, padding; container widens beyond `44rem` up to viewport cap; `overflow-x: auto`, `white-space: pre`
- Images: `max-width: 100%`, `height: auto`
- Date above title, `{{ .Content }}`, back-to-index link

## Phase 3: Build program (`cmd/build/main.go`)

### 3a. Post discovery

- `Post` struct: Title, Date, Tags, Draft, Slug, Body, Content, Excerpt, SourcePath
- `--posts-per-page` flag (default 10)
- `filepath.Glob("posts/*.md")`, read each file

### 3b. Front matter parsing

- Split YAML front matter from body via goldmark-meta
- Extract: `title` (string), `date` (string in `2006-01-02T150405` format), `draft` (bool), `tags` ([]string, optional), `slug` (string, optional)

### 3c. Slug derivation

- Use `slug` front matter if set
- Otherwise: filename minus `.md`, strip `^\d{4}-\d{2}-\d{2}T\d{6}-` then `^\d{4}-\d{2}-\d{2}-`

### 3d. Filtering and sorting

- Drop `draft: true` posts
- Sort by date descending

### 3e. Markdown rendering + syntax highlighting

- Goldmark with goldmark-highlighting: `WithStyle("monokai")`, `WithFormatOptions(chromahtml.WithClasses(true))`
- Chroma tokenizes by language from info string (` ```go `); no language → plain `<pre><code>`
- Generate Chroma CSS once via `chromahtml.New(WithClasses(true)).WriteCSS(w, styles.Get("monokai"))`, store as string for templates

### 3f. Excerpt generation

- First `<p>` from rendered HTML, strip tags, truncate ~200 chars with ellipsis

### 3g. Image processing

- Scan markdown for `![...](path)`, classify local vs remote
- Local: decode, resize to 480/768/1200/1600 (capped by source), write to `assets/{slug}/{name}-{width}w.{ext}`
- Rewrite local `<img>` with `srcset`/`sizes`; leave remote unchanged
- Add `loading="lazy"` to all `<img>`

### 3h. Page generation

- Execute `post.html.tmpl` per post → `{slug}/index.html`
- Chunk into groups of 10 → `index.html`, `page/2/index.html`, etc.

### 3i. Stale output cleanup

- Remove generated slug dirs identified by a build marker when they no longer match current non-draft posts
- Rebuild `page/` from scratch each run so excess pagination dirs disappear automatically
- Rebuild `assets/` from scratch each run so stale image variants disappear automatically

## Phase 4: Validation

Runs before rendering, fails fast on first error (report filename + field + problem):

- Front matter block exists
- `title`: present, non-empty string
- `date`: present and formatted as `YYYY-MM-DDTHHMMSS`
- `draft`: present, boolean
- `tags` (if present): list of strings
- `slug` (if present): non-empty string and not a reserved site path
- Local image refs: file exists on disk and decodes as an image
- No duplicate slugs across non-draft posts

## Phase 5: Spellchecking

Runs after validation, before rendering:

- Extract markdown body (after front matter `---`) for each non-draft post
- Run through misspell; collect filename, line, word, suggestion
- Print to stderr; abort build if any found

## Phase 6: Makefile targets

- `new` — generate `posts/YYYY-MM-DDTHHMMSS-untitled.md` with front matter template, print path
- `build` — `go run ./cmd/build`
- `publish` — alias for `build`
- `deploy` — `publish`, then `git diff --quiet` on generated output; if changed, add + commit + push; if not, print "nothing to deploy"
- `clean` — remove `index.html`, `page/*/`, slug dirs, `assets/` variants

## Phase 7: Automated tests

**Slug derivation** — front matter override; `YYYY-MM-DD-` strip; `YYYY-MM-DDTHHMMSS-` strip; no-prefix fallback

**Validation** — missing title; empty title; bad date; missing draft; non-string tag; duplicate slugs; missing local image; valid post passes

**Filtering/sorting** — drafts excluded; date-descending order

**Pagination** — 0 posts (empty index); 10 posts (one page); 11 posts (two pages); 25 posts (three pages, correct counts)

**Syntax highlighting** — language-tagged block has Chroma class spans; untagged block is plain `<pre><code>`; generated CSS contains expected selectors

**Responsive images** — local image produces width variants capped by source; `srcset`/`sizes` present; remote image unchanged; `loading="lazy"` on all `<img>`

**Stale cleanup** — deleted post removes slug dir; drafted post removes slug dir; fewer posts removes excess page dirs; removed image removes asset variants

## Phase 8: Sample post and verification

Create a post exercising: blockquote, language-tagged code block, untagged code block, long-line code block, local image, inline code.

Verify: index listing with excerpts, post page navigation, dark theme colors, narrow prose column, blockquote styling, syntax-highlighted spans on tagged blocks, plain rendering on untagged blocks, horizontal scroll on long lines, wider code container, responsive `srcset` output, `loading="lazy"`, mobile layout (~375px), pagination with multiple posts, draft exclusion, spellcheck failure on typo.

## Implementation Notes

- Single-file `cmd/build/main.go` — don't over-structure for a personal blog
- All CSS inline in templates, no external stylesheets
- Generated output committed to git; deterministic build so `git diff` detects changes
- Generated post pages include a marker comment so cleanup can distinguish blog output from other project directories in the repo

## Tasks

- Scaffold Go module, directories, and Makefile
- Build HTML templates (index + post) with dark theme and all styling
- Implement post discovery, front matter parsing, slug derivation, and validation
- Implement rendering, syntax highlighting, excerpt generation, and image processing
- Implement page generation, pagination, and stale output cleanup
- Wire up Makefile targets
- Add automated tests
- Add sample post and verify
