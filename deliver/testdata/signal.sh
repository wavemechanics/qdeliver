#!/bin/sh

f() {
    sleep 2
    kill $1
}

f "$$" &
sleep 5
exit 0