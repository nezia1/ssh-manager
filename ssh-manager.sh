#!/bin/sh

connection="$(./main)"

./main | xargs -I {} eval '{}'
