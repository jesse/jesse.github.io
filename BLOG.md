# Blog Generator

A static blog generator written in Go. It converts Markdown posts into HTML pages with responsive images, syntax highlighting, and pagination.

## Quick Start

```bash
make new        # create a new post
# edit the file printed to stdout
make build      # generate HTML
make deploy     # build, commit, and push
```

## Creating a Post

```bash
make new
```

This creates a file at `posts/YYYY-MM-DDTHHMMSS-untitled.md` with starter frontmatter. Rename the file to something descriptive -- the filename becomes the URL slug (minus the date prefix).

For example, `posts/2026-04-09T110000-launching-the-blog.md` produces the slug `launching-the-blog` and is served at `/launching-the-blog/`.

## Post Format

Each post is a Markdown file in `posts/` with YAML frontmatter:

```yaml
---
title: "My Post Title"
date: "2026-04-09T110000"
tags:
  - go
  - blog
draft: false
---

Post content in Markdown goes here.
```

### Frontmatter Fields

| Field   | Required | Description                                                    |
|---------|----------|----------------------------------------------------------------|
| `title` | Yes      | Post title. Must be non-empty for published posts.             |
| `date`  | Yes      | Publish date in `YYYY-MM-DDTHHMMSS` format (no colons).       |
| `tags`  | No       | List of string tags.                                           |
| `draft` | Yes      | `true` to hide the post from the site, `false` to publish it. |
| `slug`  | No       | Override the auto-derived slug from the filename.              |

### Content

Standard Markdown is supported (GitHub Flavored Markdown via goldmark):

- **Code blocks** with a language tag get syntax highlighting (monokai theme). Plain code blocks without a language tag are rendered as scrollable preformatted text.
- **Blockquotes**, inline code, links, and all other standard Markdown features work as expected.
- **Images** can be local or remote. Local images placed in the `posts/` directory are automatically resized into responsive variants (480w, 768w, 1200w, and original width) and served with `srcset`.

```markdown
![Alt text](my-image.png)
```

Place the image file (e.g., `my-image.png`) in the `posts/` directory alongside your Markdown file.

## Building

```bash
make build
```

This parses all posts, validates frontmatter, spell-checks published posts, renders Markdown to HTML, generates responsive image variants, and writes the output:

- Each post gets its own directory: `<slug>/index.html`
- Index pages are paginated: `index.html`, `page/2/index.html`, `page/3/index.html`, etc.
- Image variants are written to `assets/`

To change the number of posts per page:

```bash
make build POSTS_PER_PAGE=20
```

## Deploying

```bash
make deploy
```

Builds the site, then commits and pushes all generated files if there are changes. If nothing changed, it prints "nothing to deploy".

## Cleaning

```bash
make clean
```

Removes all generated HTML and asset files. Source files in `posts/`, `templates/`, and `cmd/` are never touched.

## Validation

The build fails if any of these checks fail:

- Missing or empty `title` on a published post
- Invalid `date` format (must be `YYYY-MM-DDTHHMMSS`)
- Missing `draft` field
- Tags that aren't strings
- Duplicate slugs across published posts
- Slug contains `/`, `\`, or starts with `.`
- Slug collides with a reserved directory name
- Local image referenced in a post doesn't exist in `posts/`
- Spelling errors in published posts (via misspell)

## Project Structure

```
posts/              Markdown source files and local images
templates/          HTML templates (index.html.tmpl, post.html.tmpl)
cmd/build/          Go source for the generator
assets/             Generated responsive image variants
page/               Generated paginated index pages
<slug>/index.html   Generated individual post pages
```
