#!/bin/bash

version=$1

if [ -z "$version" ]
then
  echo "Usage: $0 <version>"
  exit
fi

WORKING_DIR=godbledger-$version

echo "Working in $WORKING_DIR..."

mkdir -p $WORKING_DIR
cd $WORKING_DIR

tar -czvf godbledger-$build-x64-$version.tar.gz godbledger-$build-x64-$version

echo '#### sha256sum'
sha256sum godbledger-*-x64-$version.zip
