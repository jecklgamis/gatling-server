#!/usr/bin/env bash
set -e
EXEC_DIR="$(pwd)"

GIT_TAG=$(git describe --tags --exact-match 2>/dev/null || true)
if [ -n "${GIT_TAG}" ]; then
  VERSION="${GIT_TAG#v}"
  echo "Using version ${VERSION} from git tag ${GIT_TAG}"
else
  VERSION="0.0.0-dev+$(git rev-parse --short HEAD)"
  echo "HEAD is not exactly tagged, using dev version ${VERSION}"
fi

TARGET_OS=${TARGET_OS:-linux}
TARGET_ARCH=${TARGET_ARCH:-amd64}
TARGET_PLATFORM=${TARGET_OS}-${TARGET_ARCH}

BASE_DIST_DIR=dist
BASE_DIR_NAME=gatling-server-${TARGET_PLATFORM}-${VERSION}
DIST_DIR=${BASE_DIST_DIR}/${BASE_DIR_NAME}
TAR_GZ_NAME=${BASE_DIR_NAME}.tar.gz
SHA_NAME=${BASE_DIR_NAME}.sha256

echo "Creating distribution for ${TARGET_PLATFORM}"
echo "VERSION = ${VERSION}"

rm -rf ${DIST_DIR}
mkdir -p ${DIST_DIR}/bin
mkdir -p ${DIST_DIR}/configs
mkdir -p ${DIST_DIR}/scripts

cp bin/gatling-server-${TARGET_PLATFORM} ${DIST_DIR}/bin/gatling-server
cp -rf configs/* ${DIST_DIR}/configs/
cp scripts/gatling-jar-runner.sh ${DIST_DIR}/scripts/
chmod +x ${DIST_DIR}/scripts/gatling-jar-runner.sh

cat <<'EOF' >${DIST_DIR}/run-server.sh
#!/usr/bin/env bash

APP_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
export APP_ENVIRONMENT=prod
${APP_DIR}/bin/gatling-server

EOF
chmod +x ${DIST_DIR}/run-server.sh

cp server.key ${DIST_DIR}
cp server.crt ${DIST_DIR}

cd ${BASE_DIST_DIR} && tar cvzf ${TAR_GZ_NAME} ${BASE_DIR_NAME} >/dev/null 2>&1 &&
  cd ${EXEC_DIR} && echo "Wrote ${BASE_DIST_DIR}/${TAR_GZ_NAME}"
cd ${BASE_DIST_DIR} && shasum -a 256 ${TAR_GZ_NAME} | awk '{ print $1 }' >${SHA_NAME} && cd ${EXEC_DIR} && echo "Wrote ${BASE_DIST_DIR}/${SHA_NAME}"
