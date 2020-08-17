#!/bin/bash

this=$(pwd)

cat submodules | while read submodule
do
  printf 'Updateing %s module: %s\n' "$submodule" "$1"
  cd "$submodule"
  go get -u "$1"
  cd "$this"
done

