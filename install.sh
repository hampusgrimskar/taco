#!/bin/bash

set -e

git clone https://github.com/hampusgrimskar/taco.git

cd taco

cmake . && make

if [ $? -ne 0 ]; then
    echo "Error: Failed to build taco, make sure you have installed the required dependencies\n Please refer to https://github.com/hampusgrimskar/taco/blob/master/README.md if you are unsure."
    exit 1
fi

echo "export PATH=\"\$PATH:$pwd/taco/bin\"" >> ~/.bashrc