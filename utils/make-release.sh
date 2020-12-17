#!/bin/bash

set -e

version=$1

if [ -z "$version" ]
then
  echo "Usage: $0 <version>"
  exit
fi

if [ "$release_pattern" == "xgo" ]; then
  # to build all: make VERSION=$version build-cross
  make VERSION=$version build-linux-amd64
  make VERSION=$version build-linux-arm-7
  make VERSION=$version build-linux-arm64

  WORKING_DIR=release/
  echo "Working in $WORKING_DIR..."
  mkdir -p $WORKING_DIR
  cd $WORKING_DIR

  tar -czvf godbledger-linux-x64-v$version.tar.gz -C ../build/dist/linux-amd64 .
  tar -czvf godbledger-arm7-v$version.tar.gz  -C ../build/dist/linux-arm-7 .
  tar -czvf godbledger-arm64-v$version.tar.gz -C ../build/dist/linux-arm64 .
else
  make VERSION=$version linux
  make VERSION=$version linux-arm-7
  make VERSION=$version linux-arm-64

  WORKING_DIR=release/
  echo "Working in $WORKING_DIR..."
  mkdir -p $WORKING_DIR
  cd $WORKING_DIR

  tar -czvf godbledger-linux-x64-v$version.tar.gz godbledger-linux-x64-v$version
  tar -czvf godbledger-arm7-v$version.tar.gz godbledger-arm7-v$version
  tar -czvf godbledger-arm64-v$version.tar.gz godbledger-arm64-v$version
fi

echo '#### sha256sum'
sha256sum godbledger-*-v$version.tar.gz
