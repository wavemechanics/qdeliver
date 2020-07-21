#!/bin/sh

cat <<EOF > "$TESTDIR/notify.out"
recipient: $1
newaddress: $2
EOF
