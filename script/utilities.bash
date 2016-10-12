#!/bin/bash
#
# Utility functions for pretty output

cd "$(dirname "$0")"
cd ..

sent_header=f
colors=$(( $(tput colors 2> /dev/null || :) + 0 ))

function colorize()
{
  if (( colors >= 8 )); then
    tput bold
    tput setaf 4
  fi
  cat -
  if (( colors >= 8 )); then
    tput sgr0
  fi
}

function header()
{
  if [ "$sent_header" = t ]; then
    echo
  fi
  echo "$*" | colorize
  echo '----------------------------------------------------------------' | colorize
  sent_header=t
}

function join_by()
{
  local IFS="$1"; shift; echo "$*";
}

# vim: set ft=sh :
