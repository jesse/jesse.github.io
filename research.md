# Blog Research

## Goal

A static blog on GitHub Pages. Dark theme, no JavaScript, responsive. Posts authored in markdown, validated and spellchecked, built to static HTML with a Go program.

## Design

- Dark background (`#1a1a1a`), light text (`#e8e8e8`), light blue links (`#5599ff`)
- Serif body font (Georgia), generous `line-height`, centered prose column around `44rem` (~704px), inspired by the narrow single-column reading feel of DHH's HEY World page at `https://world.hey.com/dhh` (reviewed April 9, 2026)
- Responsive via `max-width` + `padding`
- System fonts only — no external loading
- Blockquotes should read like intentional pull-quotes/callouts: accent rule, subtle surface tint, generous spacing, and slightly softer text color
- Fenced code blocks should use syntax highlighting, a dedicated monospace stack, and a darker surface
- Syntax highlighting pipeline: Chroma lexer tokenizes the code (language detected from the fenced code block info string, e.g., ` ```go `), a formatter emits HTML spans, and a style provides the color palette
- Use **CSS classes mode** (`WithClasses(true)`) rather than inline styles — generate the CSS once at build time via `formatter.WriteCSS()` and embed it in the template `<style>` tag. This keeps HTML smaller and avoids redundant `style=` attributes on every span
- Use a dark-friendly Chroma style (e.g., `monokai`, `dracula`, or `github-dark`) that pairs well with the `#1a1a1a` page background
- The highlighting extension wraps output in `<div class="highlight"><pre ...><code>...</code></pre></div>` with token spans like `<span class="k">if</span>` for keywords
- Large code blocks should be allowed to exceed the prose column up to a wider viewport cap, then scroll horizontally when lines are still too long
- `loading="lazy"` on all `<img>` tags
- Local images should be resized into a small set of responsive widths and emitted with `srcset`/`sizes`; remote images are left unchanged

## URL Structure

- **Index**: `/` (page 1), `/page/2/`, `/page/3/`, etc. — 10 posts per page
- **Posts**: `/{slug}/` — slug-only, no date in URL (e.g., `/panther-lake-is-the-real-deal/`)
- Slug derived from filename (strip `YYYY-MM-DD-` or `YYYY-MM-DDTHHMMSS-` prefix) or `slug` front matter field
- Clean URLs via GitHub Pages serving `{dir}/index.html` as `{dir}/`

## Page Layouts

**Index** — single-column feed, newest first. Each entry: date → title (linked) → excerpt. "Projects" section in header with existing links (Math Adventures, Waveform). Paginated with "Older/Newer posts" links.

**Post** — date above title, full rendered content below, link back to index. Same theme as index.

## Toolchain

| Concern | Tool |
|---------|------|
| Markdown → HTML | goldmark + goldmark-meta (Go) |
| Syntax highlighting | Chroma via `github.com/yuin/goldmark-highlighting/v2` |
| Spellcheck | misspell (Go) |
| Image resizing | Go stdlib + `golang.org/x/image/draw` |
| Validation | Custom Go (front matter, date, title, formatting, local image refs) |
| Build | `cmd/build/main.go` |
| Orchestration | Makefile wrapping `go run` |

All Go-based, no npm.

## Build Workflow

- `make new` — creates `posts/YYYY-MM-DDTHHMMSS-untitled.md` with front matter template:
  ```yaml
  ---
  title: ""
  date: "YYYY-MM-DDTHHMMSS"
  tags: []
  draft: false
  ---
  ```
- The builder accepts only `YYYY-MM-DDTHHMMSS` for the `date` field
- `make build` — validates, spellchecks, syntax-highlights code, generates responsive local image variants, renders all non-draft posts, outputs index + post pages
- `make publish` — runs validation + spellcheck + full rebuild
- `make deploy` — runs `make publish`, then checks whether generated site output changed; if it did, commits and pushes; if not, prints "nothing to deploy" and exits
- Posts with `draft: true` are skipped (unpublish without deleting)
- Every deploy rebuilds all posts from scratch — no incremental build or change detection needed. Build is deterministic and fast. Git diff tells us if anything actually changed.

## File Structure

```
cmd/build/main.go               # build program
posts/2026-04-09T110000-first-post.md  # markdown source
templates/
  index.html.tmpl               # index page template (listing + pagination)
  post.html.tmpl                # individual post template
Makefile
go.mod / go.sum

# Generated output:
index.html                      # index page 1
page/2/index.html               # index page 2, etc.
{slug}/index.html               # individual post pages
assets/{slug}/hero-768w.jpg     # generated responsive image variant
```

## Authoring Workflow

1. **Create** — `make new` scaffolds a new post in `posts/` with front matter template and prints the created path
2. **Write** — author the post in markdown, fill in title/tags, add images/video as needed
3. **Preview** — `make build` generates the site locally; open `index.html` in a browser to review
4. **Publish** — `make deploy` validates, spellchecks, builds, commits generated HTML, and pushes to GitHub Pages. Skips commit/push if nothing changed.
5. **Unpublish** — set `draft: true` in front matter, run `make deploy`

What `make publish` does under the hood:
- Validates all non-draft posts (front matter fields, date format, non-empty title)
- Runs spellcheck on post content
- Renders markdown to HTML via goldmark
- Applies syntax highlighting to fenced code blocks
- Generates responsive width variants for local images and emits `srcset`/`sizes`
- Generates index pages (with pagination) and individual post pages
- Fails fast on validation or spellcheck errors — nothing is built if a check fails

## Decisions

- Existing project links grouped under "Projects" in header/nav
- The artifacts in this directory describe the intended implementation scope; they are not split into separate version buckets
- Use a prose column around `44rem`, borrowing the narrow, readable feel of DHH's HEY World layout without trying to copy it pixel-for-pixel
- Style blockquotes as accented callouts rather than leaving browser defaults in place
- Use syntax highlighting for fenced code blocks and let code containers extend wider than prose text before falling back to horizontal scrolling
- Generate responsive width variants for local images during the build and emit `srcset`/`sizes`; leave remote image URLs unchanged
- Local images referenced from markdown should use relative paths and be emitted to `assets/{slug}/...` in the generated site
- Use a marker comment in generated post pages so cleanup only removes blog-owned slug directories; rebuild `page/` and `assets/` from scratch on each publish/build
- Keep `make new` environment-agnostic: scaffold the markdown file and print its path instead of trying to open an editor automatically.

## Problems to Solve

### Main problem

Build a deterministic, no-JavaScript static blog generator in Go that turns markdown posts into a GitHub Pages site with validation, spellchecking, pagination, strong typography, and first-class rendering for quotes, code, and images.

### 1. Define the content contract

- Decide the source filename convention for posts created by `make new`
- Define the front matter schema: required fields, optional fields, and types
- Define publishability rules: which posts are included or skipped
- Define URL rules so slugs and output paths are deterministic

Elementary problems:

- `make new` should create `posts/YYYY-MM-DDTHHMMSS-untitled.md`
- Required front matter fields are `title`, `date`, and `draft`
- Optional front matter fields are `tags` and `slug`
- `draft: true` excludes a post from generated output
- `slug` front matter wins; otherwise derive the slug from the filename by removing a leading `YYYY-MM-DD-` or `YYYY-MM-DDTHHMMSS-` prefix

### 2. Load and validate source posts

- Discover post files under `posts/`
- Split front matter from markdown body
- Parse metadata into typed values
- Reject malformed or ambiguous content before any rendering happens

Elementary problems:

- Read every `.md` file in `posts/`
- Parse YAML front matter and markdown body separately
- Ensure `title` is a non-empty string
- Ensure `date` is parseable and sortable
- Ensure `draft` is a boolean
- Ensure `tags`, if present, is a list of strings
- Ensure `slug`, if present, is a non-empty string
- Ensure `slug`, if present, does not collide with reserved site directories
- Detect missing local images referenced from markdown
- Detect duplicate slugs across non-draft posts
- Report validation errors with filename, field, and problem

### 3. Define the presentation system

- Choose the body copy width and page spacing
- Decide how blockquotes should stand out from normal paragraphs
- Decide how code should render inside a narrow reading layout
- Decide how images should behave inside posts

Elementary problems:

- Use a prose column around `44rem`
- Let code blocks expand beyond the prose column up to a wider viewport cap
- Give blockquotes an accent rule, subtle surface, and generous spacing
- Style the `.highlight > pre` wrapper with a darker surface than the page background, a dedicated monospace font stack, and padding
- Allow the code container to grow wider than the `44rem` prose column up to a viewport cap, then `overflow-x: auto` for horizontal scroll on long lines
- Embed the generated Chroma CSS (token color rules like `.chroma .k { color: ... }`) in the template `<style>` block; the build program passes this CSS as a template variable
- Keep inline code distinct but understated (subtle background, monospace, no highlighting)
- Ensure images scale to the content width without breaking layout

### 4. Transform posts and local media

- Convert markdown into HTML
- Produce index-friendly excerpts
- Apply lightweight HTML post-processing required by the site
- Generate responsive image assets for local media
- Order posts consistently for output generation

Elementary problems:

- Render markdown body to HTML with Goldmark
- Register `github.com/yuin/goldmark-highlighting/v2` as a goldmark extension
- Choose a dark-friendly Chroma style (e.g., `monokai`) that pairs well with the `#1a1a1a` page background
- Configure with `WithClasses(true)` so Chroma emits CSS-class spans (e.g., `<span class="k">if</span>`) instead of inline `style=` attributes — keeps HTML smaller
- At build time, call `chromahtml.New(chromahtml.WithClasses(true)).WriteCSS(w, styles.Get("monokai"))` once to generate the highlight stylesheet and pass it to templates
- Chroma detects the language from the fenced block info string (e.g., ` ```go `, ` ```python `) and tokenizes into token types (Keyword, String, Comment, etc.)
- Blocks without a language tag render as plain `<pre><code>` with no highlighting
- Add `loading="lazy"` to generated `<img>` tags
- For local image references, generate a small width set such as `480`, `768`, `1200`, and `1600`, capped by source width
- Emit `srcset` and `sizes` for local images
- Leave remote images unchanged
- Extract the first meaningful paragraph as excerpt text
- Sort published posts by date descending

### 5. Generate the site output

- Render individual post pages
- Render paginated index pages
- Keep output paths compatible with GitHub Pages clean URLs
- Remove obsolete generated files when posts are deleted or unpublished

Elementary problems:

- Write each post to `{slug}/index.html`
- Write page 1 to `index.html`
- Write later pages to `page/N/index.html`
- Put shared dark-theme, responsive styling in templates, including blockquote and code block rules
- Keep wide code blocks usable on mobile and desktop
- Remove stale post directories for deleted or drafted posts
- Remove stale pagination directories when page count shrinks
- Remove stale generated image variants that are no longer referenced
- Keep cleanup scoped to blog-generated output even though the repo also contains other project directories

### 6. Support the author workflow

- Make content creation fast
- Make publish/deploy commands predictable
- Keep deploy behavior safe and deterministic

Elementary problems:

- `make new` scaffolds a correctly named post with front matter template
- `make build` performs validation, spellcheck, and full rendering
- `make publish` remains the canonical pre-deploy build command
- `make deploy` only commits and pushes when generated output changed

### 7. Verify correctness

- Cover the core logic with automated tests
- Keep at least one manual smoke-check path for rendered output

Elementary problems:

- Add tests for slug derivation
- Add tests for front matter validation failures, duplicate slugs, and missing local images
- Add tests for draft exclusion
- Add tests for pagination boundaries
- Add tests for syntax-highlighted code block rendering and overflow-friendly wrappers
- Add tests for responsive image variant generation and emitted `srcset`/`sizes`
- Add tests for stale output cleanup, including generated image variants
- Add a sample post or fixture for manual visual verification that exercises blockquotes, code blocks, and images
