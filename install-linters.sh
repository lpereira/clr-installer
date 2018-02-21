#!/bin/bash

echo "installing gometalinter.v2"
if [[ ! $(type -P "gometalinter.v2") ]]; then
    go get -v -u gopkg.in/alecthomas/gometalinter.v2
fi

echo "installing ineffassign"
if [[ ! $(type -P "ineffassign") ]]; then
    go get -v -u github.com/gordonklaus/ineffassign
fi

echo "installing unused"
if [[ ! $(type -P "unused") ]]; then
    go get -v -u honnef.co/go/tools/cmd/unused
fi

echo "installing errcheck"
if [[ ! $(type -P "errcheck") ]]; then
    go get -v -u github.com/kisielk/errcheck
fi

echo "installing misspell"
if [[ ! $(type -P "misspell") ]]; then
    go get -v -u github.com/client9/misspell/cmd/misspell
fi

echo "installing deadcode"
if [[ ! $(type -P "deadcode") ]]; then
    go get -v -u github.com/tsenart/deadcode
fi

echo "installing structcheck"
if [[ ! $(type -P "structcheck") ]]; then
    go get -v -u github.com/opennota/check/cmd/structcheck
fi

echo "installing varcheck"
if [[ ! $(type -P "varcheck") ]]; then
    go get -v github.com/opennota/check/cmd/varcheck
fi

echo "installing goling"
if [[ ! $(type -P "golint") ]]; then
    go get -v -u github.com/golang/lint/golint
fi

echo "installing govet"
if [[ ! $(type -P "govet") ]]; then
    mkdir -p ~/go/bin
    touch ~/go/bin/govet
    chmod 755 ~/go/bin/govet

    cat <<EOF > ~/go/bin/govet
#!/bin/bash
go vet \$@
EOF
fi
