#!/bin/bash

set -e

sudo apt update

sudo apt install git -y
sudo apt install g++ -y
sudo apt install cmake -y
sudo apt install tmux -y
sudo apt install libncurses5-dev libncursesw5-dev -y

git clone https://github.com/hampusgrimskar/taco.git

cd taco

cmake . && make

# curl -o- https://raw.githubusercontent.com/hampusgrimskar/taco/refs/heads/master/install.sh | bash
