#!/bin/sh -e

VER=$1
BRANCH=$2
PROJ="vulcand"
RELEASE_DIR="/tmp/release"
PROJ_DIR=${GOPATH}/src/github.com/mailgun/${PROJ}
RELEASE_DESCRIPTION=$3

if [ $# -lt 3 ]; then
	echo "Usage: ${0} VERSION TAG/BRANCH COMMENT" >> /dev/stderr
	exit 255
fi

set +x
set -u

function setup_env {
	local proj=${1}
	local ver=${2}
    local proj_dir=${3}

    mkdir -p $(dirname $proj_dir)

	if [ ! -d ${proj_dir} ]; then
		git clone https://github.com/mailgun/${proj} ${proj_dir}
	fi

	pushd ${proj_dir}
		git fetch --all
		git reset --hard origin/master
		git checkout $BRANCH
        git tag $ver
        git push --tags --repo https://$USERNAME:$PASSWORD@github.com/mailgun/$proj
	popd
}


function package {
	local target=${1}
	local srcdir=${2}

	local ccdir="${srcdir}/${GOOS}_${GOARCH}"
	if [ -d ${ccdir} ]; then
		srcdir=${ccdir}
	fi
	for bin in vulcand vctl vbundle; do
		cp ${srcdir}/${bin} ${target}
	done

	cp ${PROJ_DIR}/README.md ${target}/README.md
    cp ${PROJ_DIR}/CHANGELOG.md ${target}/CHANGELOG.md
}

function main {
	setup_env ${PROJ} ${VER} ${PROJ_DIR}

	for os in linux; do
		export GOOS=${os}
		export GOARCH="amd64"

		pushd $PROJ_DIR
			make install
		popd

		TARGET="${RELEASE_DIR}/vulcand-${VER}-${GOOS}-${GOARCH}"
		mkdir -p ${TARGET}
		package ${TARGET} "${GOPATH}/bin"
        cd $RELEASE_DIR

		if [ ${GOOS} == "linux" ]; then
            ARTIFACT="${TARGET}.tar.gz"
			tar cfz ${TARGET}.tar.gz $(basename ${TARGET})
			echo "Wrote ${TARGET}.tar.gz"
		else
            ARTIFACT="${TARGET}.zip"
			zip -qr ${TARGET}.zip $(basename ${TARGET})
			echo "Wrote ${TARGET}.zip"
		fi

        github-release release \
          --user mailgun \
          --repo vulcand \
          --tag $VER \
          --name $VER \
          --description $RELEASE_DESCRIPTION \
          --pre-release

       github-release upload \
          --user mailgun \
          --repo vulcand \
          --tag $VER \
          --name $(basename $ARTIFACT) \
          --file $ARTIFACT
	done
}

main
