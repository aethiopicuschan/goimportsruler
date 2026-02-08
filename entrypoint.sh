#!/bin/sh
set -eu

# If no args were passed, default to "."
if [ "$#" -eq 0 ]; then
  set -- "."
fi

exec goimportsruler "$@"
