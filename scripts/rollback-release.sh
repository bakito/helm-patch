#/bin/bash

git tag -d v${1}
git push origin :v${1}