#!/bin/sh

forward() {
    if test -z "$1"
    then
        echo "$0: forward: email addresses required" 1>&2
        exit 111
    fi
    exec /var/qmail/bin/forward "$@"
    exit 111
}

bounce() {
    echo "${1:-This address no longer accepts mail.}" 1>&2
    exit 100
}

drop() {
    echo "dropping" 1>&2
    exit 0
}

match_subject() {
    pattern=$1
    msg=$2

    sed -n '
        1,/^$/ {
        /^[sS][uU][bB][jJ][eE][cC][tT]:[ 	]*/s///p
    }' | grep -q -F "$pattern"
    
    case $? in
    0)  # pattern found in subject
        exit 0  # allow deliveries to continue
        ;;
    1)  # pattern not found in subject
        if test -n "$msg"
        then
            echo "$msg" 1>&2
        fi
        exit 100    # bounce
        ;;
    *)  # error
        exit 111
        ;;
    esac
}

main() {
    action=$1
    shift

    case $action in
    forward)
        forward "$@"
        ;;
    bounce)
        bounce "$@"
        ;;
    drop)
        drop
        ;;
    match-subject)
        match_subject "$@"
        ;;
    *)
        echo "$0: $action: undefined action" 1>&2
        exit 111
    esac
}

main "$@"
