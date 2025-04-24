apt update

apt install git -y
apt install g++ -y
apt install cmake -y
apt install tmux -y
apt install libncurses5-dev libncursesw5-dev -y

git clone https://github.com/hampusgrimskar/taco.git

cd taco

cmake . && make