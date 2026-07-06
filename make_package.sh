#!/bin/bash

goVersion=$1
CURRENT_PATH=`pwd`
export ENV_CODE_PATH="${CURRENT_PATH}"
PACKAGE_ROOT="/opt/omspkg"

die() {
    echo $1
    exit 1
}

usage() {
  script_name=`basename "$0"`
  echo "./${script_name} <version_tag>"
  exit 0
}

compile_code() {
    cd ${ENV_CODE_PATH}/main
    go env -w GO111MODULE=on
    go env -w GOPROXY=https://goproxy.cn
    go mod tidy
    chmod +x main.go
    go build -ldflags "-X 'main.Version=${goVersion}'" -o agent main.go
    [ -f agent ] || die "go build failed"

    cd ${CURRENT_PATH}
}

package_all() {
    BINARY_FILENAME="monitor_agent_setup.tar.gz"
    PKG_PATHNAME="monitor_agent_setup"
    if [ ! -z "${goVersion}" ]; then
        BINARY_FILENAME="monitor_agent_setup_${goVersion}.tar.gz"
        PKG_PATHNAME="monitor_agent_setup_${goVersion}"
    fi

    [ -f ${BINARY_FILENAME} ] && rm -f ${BINARY_FILENAME}
    [ -d ${PKG_PATHNAME} ] && rm -rf ${PKG_PATHNAME}
    mkdir -p ${PKG_PATHNAME}/bin
    mkdir -p ${PKG_PATHNAME}/bin/conf
    mkdir -p ${PKG_PATHNAME}/scripts
    mkdir -p ${PKG_PATHNAME}/3rd

    #bin
    cp ${ENV_CODE_PATH}/main/agent ${PKG_PATHNAME}/bin
    #conf
    cp ${ENV_CODE_PATH}/main/conf/agent.yaml ${PKG_PATHNAME}/bin/conf
    cp ${ENV_CODE_PATH}/main/conf/redis.conf ${PKG_PATHNAME}/bin/conf
    echo ${goVersion} >> ${PKG_PATHNAME}/bin/conf/agent_version.txt
    #shell
    dos2unix ${ENV_CODE_PATH}/main/scripts/*
    cp ${ENV_CODE_PATH}/main/scripts/* ${PKG_PATHNAME}/scripts && chmod a+x ${PKG_PATHNAME}/scripts/*

    #3rd
    cp ${ENV_CODE_PATH}/main/3rd/redis.tar.gz ${PKG_PATHNAME}/3rd

    cp ${ENV_CODE_PATH}/install.sh ${PKG_PATHNAME}
    dos2unix ${PKG_PATHNAME}/*.sh && chmod a+x ${PKG_PATHNAME}/*.sh

    tar -czvf ${BINARY_FILENAME} ${PKG_PATHNAME}
}

if [ "X$1" == "X" ]; then
    usage
fi
compile_code
package_all
