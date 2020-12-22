#!/bin/bash

set -e

version=$1

if [ -z "$version" ]
then
  echo "Usage: $0 <version>"
  exit
fi

#make release

cd build/dist

for D in *; do
    if [ -d "${D}" ]; then
        echo "${D}"   # your processing here
        tar -czvf ${D}-v$version.tar.gz ${D}/
    fi
done

echo '#### sha256sum' >> release-notes.txt
sha256sum *-v$version.tar.gz >> release-notes.txt


find . -name "*.gz" -o -name "*.txt" | tar -czvf godbledger-release-v$version.tar.gz -T -

