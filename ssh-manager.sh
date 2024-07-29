#!/bin/sh

connection="$(./main)"

[ -n "$connection" ] && eval "$connection"
