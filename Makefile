.NOTPARALLEL:

top_srcdir = $(abspath .)
MAKEFLAGS += -r --no-print-directory

build_dir = $(top_srcdir)/build
build_bin_dir = $(build_dir)/bin
pkg_dir = $(top_srcdir)

export GOPATH=$(pkg_dir)
export GO_PACKAGE_PREFIX := clr-installer

build:
	go get -v ${GO_PACKAGE_PREFIX}/clr-installer-tui
	go install -v ${GO_PACKAGE_PREFIX}/clr-installer-tui
	go get -v ${GO_PACKAGE_PREFIX}/clr-installer-gui
	go install -v ${GO_PACKAGE_PREFIX}/clr-installer-gui

check:
	go test -cover ${GO_PACKAGE_PREFIX}/...

PHONY += coverage
coverage: build
	@rm -rf .coverage/; \
	mkdir -p .coverage/; \
	for pkg in $$(go list $$GO_PACKAGE_PREFIX/...); do \
		file=".coverage/$$(echo $$pkg | tr / -).cover"; \
		go test -covermode="count" -coverprofile="$$file" "$$pkg"; \
	done; \
	echo "mode: count" > .coverage/cover.out; \
	grep -h -v "^mode:" .coverage/*.cover >>".coverage/cover.out"; \

PHONY += coverage-func
coverage-func: coverage
	@go tool cover -func=".coverage/cover.out"

PHONY += coverage-html
coverage-html: coverage
	@go tool cover -html=".coverage/cover.out"

PHONY += lint
lint:
	@gometalinter.v2 --deadline=10m --tests --vendor --disable-all \
	--enable=misspell \
	--enable=vet \
	--enable=ineffassign \
	--enable=gofmt \
	$${CYCLO_MAX:+--enable=gocyclo --cyclo-over=$${CYCLO_MAX}} \
	--enable=golint \
	--enable=deadcode \
	--enable=varcheck \
	--enable=structcheck \
	--enable=unused \
	--enable=vetshadow \
	--enable=errcheck \
	./...

all: build

PHONY += all

.PHONY = $(PHONY)
.DEFAULT_GOAL = all
