#!/usr/bin/env bash

set -x

#
# Update the database
#
if [ ! -d "data" ]; then
  git clone https://github.com/vulndb/data.git
fi

cd data
git pull
cd ..

mkdir -p vulndb/db/
cp -rf data/db/*.json vulndb/db/

cd data
dbhash1=$(<../vulndb/db-version.txt)
dbhash2=`git rev-parse HEAD`
if [ "$dbhash1" == "$dbhash2" ]; then
  echo "Vulndb is up to date"
  exit 0
fi
echo "$dbhash2" > ../vulndb/db-version.txt
cd ..

# Bump the version numbers
./tools/semver.sh bump patch


go-bindata -o="./bindata/bindata.go" -pkg="bindata" ./vulndb/...
# format bindata file
goimports -w "./bindata/bindata.go"
# Push to repo
git commit vulndb/version.txt vulndb/db-version.txt bindata/bindata.go -m 'Updated vulnerability database'
git push