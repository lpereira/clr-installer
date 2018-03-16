.NOTPARALLEL:

top_srcdir = $(abspath .)
MAKEFLAGS += -r --no-print-directory

build_dir = $(top_srcdir)/build
build_bin_dir = $(build_dir)/bin
pkg_dir = $(top_srcdir)
cov_dir = $(top_srcdir)/.coverage

orig_go_path = $(shell go env GOPATH)
export GOPATH=$(pkg_dir)
export GO_PACKAGE_PREFIX := clr-installer
export TESTS_DIR := $(top_srcdir)/tests/

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

PHONY += install-linters
install-linters:
ifneq ($(shell gometalinter.v2 --version 2>/dev/null 1>&2 ; echo $$?),0)
	@echo "Installing linters..."
	@GOPATH=${orig_go_path} go get -u gopkg.in/alecthomas/gometalinter.v2 1>/dev/null
	@GOPATH=${orig_go_path} gometalinter.v2 --install 1>/dev/null
endif

PHONY += update-linters
update-linters:
ifeq ($(shell gometalinter.v2 --version 2>/dev/null 1>&2 ; echo $$?),0)
	@echo "Updating linters..."
	@GOPATH=${orig_go_path} gometalinter.v2 --update 1>/dev/null
else
	@echo "Linters not installed"
	@exit 1
endif

PHONY += lint
lint: install-linters
	@gometalinter.v2 --deadline=10m --tests --vendor \
	--exclude=vendor --disable-all \
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

PHONY += install-dep
install-dep:
ifneq ($(shell dep version 2>/dev/null 1>&2 ; echo $$?),0)
	@echo "Installing dep..."
	@mkdir -p ${orig_go_path}/bin
	@curl https://raw.githubusercontent.com/golang/dep/master/install.sh 2>/dev/null \
		| GOPATH=${orig_go_path} bash
endif

PHONY += update-dep
update-dep:
ifeq ($(shell dep version 2>/dev/null 1>&2 ; echo $$?),0)
	@echo "Updating dep..."
	@curl https://raw.githubusercontent.com/golang/dep/master/install.sh 2>/dev/null \
		| GOPATH=${orig_go_path} bash
else
	@echo "Dep not installed"
	@exit 1
endif

PHONY += update-vendor
update-vendor: install-dep
	@cd ${GOPATH}/src/${GO_PACKAGE_PREFIX} ; dep ensure -update

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
