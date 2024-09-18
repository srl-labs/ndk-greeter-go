#!/usr/bin/env bash

set -o errexit
set -o pipefail

# abs path to the directory that hosts the run.sh script
BASE_DIR=$(dirname "$(readlink -f "$0")")
APPNAME=greeter
GOPKGNAME=${APPNAME}
BIN_DIR=${BASE_DIR}/build
BINARY=${BASE_DIR}/build/${APPNAME}
LABDIR=${BASE_DIR}/lab
LABFILE=${APPNAME}.clab.yml


COMMON_LDFLAGS="-X main.version=dev -X main.commit=$(git rev-parse --short HEAD)"

if [ -z "$NDK_DEBUG" ]; then
    # when not in debug mode use linker flags -s -w to strip the binary
    LDFLAGS="-ldflags=\"-s -w $COMMON_LDFLAGS\""
else
    # when NDK_DEBUG is set
    LDFLAGS="-ldflags=\"$COMMON_LDFLAGS\""
    GCFLAGS="-gcflags=\"all=-N -l\""

    # links the dlv binary to the debug directory as a hardlink
    # making it available to the greeter container when running in debug mode.
    ln -f $(which dlv) ${BASE_DIR}/debug/
fi

#################################
# Build functions
#################################
function lint-yang {
    echo "Linting YANG files"
    docker run --rm -v ${BASE_DIR}:/work ghcr.io/hellt/yanglint yang/*.yang
}

function lint-yaml {
    echo "Linting YAML files"
    docker run --rm -v ${BASE_DIR}/${APPNAME}.yml:/data/${APPNAME}.yml cytopia/yamllint -d relaxed .
}

function lint {
    lint-yang
    lint-yaml
}

GOFUMPT_CMD="docker run --rm -it -e GOFUMPT_SPLIT_LONG_LINES=on -v ${BASE_DIR}:/work ghcr.io/hellt/gofumpt:0.3.1"
GOFUMPT_FLAGS="-l -w ."

GODOT_CMD="docker run --rm -it -v ${BASE_DIR}:/work ghcr.io/hellt/godot:1.4.11"
GODOT_FLAGS="-w ."

function gofumpt {
    ${GOFUMPT_CMD} ${GOFUMPT_FLAGS}
}

function godot {
    ${GODOT_CMD} ${GODOT_FLAGS}
}

function format {
    gofumpt
    godot
}

function build-app {
    lint
    format
    echo "Building application"
    mkdir -p ${BIN_DIR}
    go mod tidy
    CMD="go build ${LDFLAGS} ${GCFLAGS} -race -o ${BINARY} ."
    echo "Executing: $CMD"
    eval $CMD
}

#################################
# High-Level run functions
#################################
function deploy-all {
    check-clab-version
    template-files
    build-app
    deploy-lab
    install-app
}

# This function is used to re-deploy the app
# without re-deploying the lab
# The workflow is:
# 1. first deploy the lab with `./run.sh deploy-all`
# 2. make changes to the app code
# 3. run `./run.sh build-restart-app`
# which will rebuild the app and restart it without
# requiring to re-deploy the lab
function build-restart-app {
    build-app
    reload-app_mgr
    sleep 3 # wait 3s for app_mgr to reload
    restart-app
}

function template-files {
    template-lab
    template-app
}


#################################
# Lab functions
#################################
function deploy-lab {
    mkdir -p logs/srl
    mkdir -p logs/greeter
    sudo clab dep -c -t ${LABDIR}
}

function destroy-lab {
    sudo clab des -c -t ${LABDIR}/${LABFILE}
    sudo rm -rf logs/srl/* logs/greeter/*
}

function check-clab-version {
    version=$(clab version | awk '/version:/ {print $2}')
    if [[ $(echo "$version 0.48.6" | tr " " "\n" | sort -V | head -n 1) != "0.48.6" ]]; then
        echo "Upgrade containerlab to v0.48.6 or newer
        Run 'sudo containerlab version upgrade' or use other installation options - https://containerlab.dev/install"
        exit 1
    fi
}

# template lab file. When NDK_DEBUG env var is set to any value
# the debug section is added to the lab file
function template-lab {
    echo "Templating lab file"
    sudo docker run --rm -e NDK_DEBUG=${NDK_DEBUG} -v ${BASE_DIR}/lab/:/tmp/ \
        hairyhenderson/gomplate:v3.11-slim \
        --file /tmp/${LABFILE}.go.tpl -o /tmp/${LABFILE}
}

#################################
# App functions
#################################

# template app file. When NDK_DEBUG env var is set to any value
# the debug section is added to the app configuration file.
function template-app {
    echo "Templating app file"
    sudo docker run --rm -e NDK_DEBUG=${NDK_DEBUG} -e NOWAIT=${NOWAIT} \
        -v ${BASE_DIR}:/tmp \
        hairyhenderson/gomplate:v3.11-slim \
        --file /tmp/${APPNAME}.yml.go.tpl -o /tmp/${APPNAME}.yml
}

# install-app creates app symlinks and reloads app_mgr
# which effectively installs and starts the app as app_mgr
# becomes aware of it
# this technique is used so that we can re-build the app later
# and have the new binary picked up by app_mgr without redeploying the lab
function install-app {
    create-app-symlink
    reload-app_mgr
}

function show-app-status {
    sudo clab exec --label containerlab=greeter --cmd "sr_cli show system application ${APPNAME}"
}

function restart-app {
    sudo clab exec --label containerlab=greeter --cmd "sr_cli tools system app-management application ${APPNAME} restart"
}

function reload-app {
    sudo clab exec --label containerlab=greeter --cmd "sr_cli tools system app-management application ${APPNAME} reload"
}

function stop-app {
    sudo clab exec --label containerlab=greeter --cmd "sr_cli tools system app-management application ${APPNAME} stop"
}

function start-app {
    sudo clab exec --label containerlab=greeter --cmd "sr_cli tools system app-management application ${APPNAME} start"
}

function redeploy-app {
    build-app
    reload-app
}

function create-app-symlink {
    sudo clab exec --label containerlab=greeter --cmd "sudo ln -s /tmp/build/${APPNAME} /usr/local/bin/${APPNAME}"
    sudo clab exec --label containerlab=greeter --cmd "sudo ln -s /tmp/${APPNAME}.yml /etc/opt/srlinux/appmgr/${APPNAME}.yml"
}


function reload-app_mgr {
    sudo clab exec --label containerlab=greeter --cmd "sr_cli tools system app-management application app_mgr reload"
}

#################################
# Packaging functions
#################################
function compress-bin {
    rm -f build/compressed
    chmod 777 build/${APPNAME}
    docker run --rm -v $(pwd):/work ghcr.io/hellt/upx:4.0.2-r0 --best --lzma -o build/compressed build/${APPNAME}
    mv build/compressed build/${APPNAME}
}

# package packages the binary into a deb package by default
# if `rpm` is passed as an argument, it will create an rpm package
function package {
    build-app
    compress-bin
    local packager=${1:-deb}
    docker run --rm -v $(pwd):/tmp -w /tmp ghcr.io/goreleaser/nfpm:v2.40.0 package \
    --config /tmp/nfpm.yml \
    --target /tmp/build \
    --packager ${packager}
}

_run_sh_autocomplete() {
    local current_word
    COMPREPLY=()
    current_word="${COMP_WORDS[COMP_CWORD]}"

    # Get list of function names in run.sh
    local functions=$(declare -F -p | cut -d " " -f 3 | grep -v "^_" | grep -v "nvm_")

    # Generate autocompletions based on the current word
    COMPREPLY=( $(compgen -W "${functions}" -- ${current_word}) )
}

# Specify _run_sh_autocomplete as the source of autocompletions for run.sh
complete -F _run_sh_autocomplete ./run.sh

function help {
  printf "%s <task> [args]\n\nTasks:\n" "${0}"

  compgen -A function | grep -v "^_" | grep -v "nvm_" | cat -n

  printf "\nExtended help:\n  Each task has comments for general usage\n"
}

# This idea is heavily inspired by: https://github.com/adriancooney/Taskfile
TIMEFORMAT=$'\nTask completed in %3lR'
time "${@:-help}"
