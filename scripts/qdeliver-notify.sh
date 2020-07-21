#!/bin/sh

usage() {
    echo "usage: $0 <recipient> <newaddress>" 1>&2
    exit 2
}

main() {
    recipient=$1
    newaddress=$2

    if test -z "$recipient" -o -z "$newaddress"
    then
        usage
    fi

    /var/qmail/bin/qmail-inject <<EOF
From: $recipient
To: $recipient
Subject: New Address Created

A new email address was created: $newaddress

To modify the behaviour of this address, go to your webdav area for
that domain and edit the file for that address.
EOF
}

main "$@"
