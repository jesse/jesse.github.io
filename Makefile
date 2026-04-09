POSTS_PER_PAGE ?= 10

build:
	go run ./cmd/build --posts-per-page $(POSTS_PER_PAGE)

publish: build

new:
	@mkdir -p posts
	@stamp=$$(date +"%Y-%m-%dT%H%M%S"); \
	path="posts/$${stamp}.md"; \
	printf '%s\n' '---' 'title: ""' "date: \"$${stamp}\"" 'tags: []' 'draft: false' '---' '' > "$$path"; \
	printf '%s\n' "$$path"

deploy: publish
	@if git diff --quiet -- index.html page p; then \
		printf '%s\n' 'nothing to deploy'; \
	else \
		git add Makefile go.mod go.sum cmd posts templates index.html page p; \
		git commit -m "Publish blog updates"; \
		git push; \
	fi

clean:
	go run ./cmd/build --clean
