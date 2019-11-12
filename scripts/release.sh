#!/bin/sh -e

set -e


CHANGED=$(git diff-index --name-only HEAD --)
if [[ ! -z $CHANGED ]]; then
    echo "Please commit your local changes first"
    exit 1
fi

sed -i "s/version: .*/version: \"${1}\"/" plugin.yaml
git add . 
git commit -m "prepare release ${1}"
git push

git tag -a v${1} -m "release ${1}"
git push origin v${1}

goreleaser --rm-dist
