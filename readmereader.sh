#!/bin/bash
golangVersion=$(sed -n '9p' readme.md | awk -F '\|' '{gsub (" ", "", $0);print $3}')

echo $golangVersion
