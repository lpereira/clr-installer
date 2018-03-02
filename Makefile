.NOTPARALLEL:

top_srcdir = $(abspath .)
MAKEFLAGS += -r --no-print-directory

build_dir = $(top_srcdir)/build
build_bin_dir = $(build_dir)/bin
pkg_dir = $(top_srcdir)
cov_dir = $(top_srcdir)/.coverage

export GOPATH=$(pkg_dir)
export GO_PACKAGE_PREFIX := clr-installer

build:
	go get -v ${GO_PACKAGE_PREFIX}/clr-installer
	go install -v ${GO_PACKAGE_PREFIX}/clr-installer

check:
	go test -cover ${GO_PACKAGE_PREFIX}/...

PHONY += coverage
coverage: build
	@rm -rf ${cov_dir}; \
	mkdir -p ${cov_dir}; \
	for pkg in $$(go list $$GO_PACKAGE_PREFIX/...); do \
		file="${cov_dir}/$$(echo $$pkg | tr / -).cover"; \
		go test -covermode="count" -coverprofile="$$file" "$$pkg"; \
	done; \
	echo "mode: count" > ${cov_dir}/cover.out; \
	grep -h -v "^mode:" ${cov_dir}/*.cover >>"${cov_dir}/cover.out"; \

PHONY += coverage-func
coverage-func: coverage
	@go tool cover -func="${cov_dir}/cover.out"

PHONY += coverage-html
coverage-html: coverage
	@go tool cover -html="${cov_dir}/cover.out"

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

PHONY += clean
clean:
	@go clean -i -r
	@git clean -fdXq

PHONY += distclean
dist-clean: clean
ifeq ($(git status -s),)
	@git clean -fdxq
	@git reset HEAD
else
	@echo "There are pending changes in the repository!"
	@git status -s
	@echo "Please check in changes or stash, and try again."
endif

all: build

PHONY += all

.PHONY = $(PHONY)
.DEFAULT_GOAL = all
