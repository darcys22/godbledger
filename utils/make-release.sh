#!/bin/bash

version=$1
build=linux
#build=arm

if [ -z "$version" ]
then
  echo "Usage: $0 <version>"
  exit
fi

make VERSION=$version release
make VERSION=$version linux-arm-7
make VERSION=$version linux-arm-64

WORKING_DIR=release/

echo "Working in $WORKING_DIR..."

mkdir -p $WORKING_DIR
cd $WORKING_DIR

tar -czvf godbledger-$build-x64-v$version.tar.gz godbledger-$build-x64-v$version

echo '#### sha256sum'
sha256sum godbledger-*-x64-v$version.tar.gz
