#!/bin/bash
set -e

# Find the directory we exist within
DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd ${DIR}/../../build

cat << EOF
# Tools

Metrictank comes with a bunch of helper tools.

Here is an overview of them all.

This file is generated by [tools-to-doc](https://github.com/grafana/metrictank/blob/master/scripts/dev/tools-to-doc.sh)

---

EOF

for tool in mt-*; do
	echo
	echo "## $tool"
	echo
	echo '```'
	./$tool -h 2>&1 | sed 's#run at most n tests in parallel (default .*)#run at most n tests in parallel (default num-processors)#'
	echo '```'
	echo
done
