#!/usr/bin/env bash
set -e
GATLING_VERSION=3.7.3

DOWNLOAD_URL=https://repo1.maven.org/maven2/io/gatling/highcharts/gatling-charts-highcharts-bundle/${GATLING_VERSION}
DIST_NAME="gatling-charts-highcharts-bundle-${GATLING_VERSION}"
DIST_ZIP_FILE="${DIST_NAME}-bundle.zip"

if  [[ -d ${DIST_NAME} ]]; then echo  "Deleting ${DIST_NAME}";  rm -rf  ${DIST_NAME}; fi
if ! [[ -f ${DIST_ZIP_FILE} ]]; then curl -fLO ${DOWNLOAD_URL}/${DIST_ZIP_FILE}; fi
unzip ${DIST_ZIP_FILE}
