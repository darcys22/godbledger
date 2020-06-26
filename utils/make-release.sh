#!/bin/bash

version=$1
build=linux

if [ -z "$version" ]
then
  echo "Usage: $0 <version> eg ./utils/make-release 0.3.0"
  exit
fi

make VERSION=$version release

WORKING_DIR=release/

echo "Working in $WORKING_DIR..."

mkdir -p $WORKING_DIR
cd $WORKING_DIR

tar -czvf godbledger-$build-x64-v$version.tar.gz godbledger-$build-x64-v$version

echo '#### sha256sum'
sha256sum godbledger-*-x64-v$version.tar.gz
