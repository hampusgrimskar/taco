#!/bin/bash

set -e

cd $(dirname $(which taco))

cd ..

git pull -r

cmake .

make