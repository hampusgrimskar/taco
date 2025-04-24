# Taco

## Dependencies

### Before installing, make sure you have the following dependencies installed on your system

    apt update

### CMake

    apt install cmake -y

### G++ Compiler

    apt install g++ -y

### ncurses header files

    apt install libncurses5-dev libncursesw5-dev -y

### Tmux

    apt install tmux -y

## Install
Run this command in the localtion where you want to install taco. The script will:
* clone this repo
* build the project
* then add taco to the path in your bashrc (if you use a different shell you will need to do this manually).
---
    curl -o- https://raw.githubusercontent.com/hampusgrimskar/taco/refs/heads/master/install.sh | bash && source ~/.bashrc

### If you make changes to the code and wish to manually build the project you can run:

    cmake . && make
