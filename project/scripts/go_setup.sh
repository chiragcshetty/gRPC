sudo apt update
cd 
sudo rm -rf /usr/local/go

wget https://go.dev/dl/go1.20.1.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.20.1.linux-amd64.tar.gz

mkdir go
cd go
mkdir bin
cd ..


if [ -z "$GOPATH" ]
then 
echo "GOPATH and GOHOME does not exist. Adding it"
cat >> ~/.bashrc <<'EOF'
export GOPATH=$(cd && cd go && pwd)
export GOBIN=$GOPATH/bin
export PATH=$PATH:$GOBIN

export PATH=$PATH:/usr/local/go/bin
EOF
source ~/.bashrc   # learnt later: this does not execute because: https://askubuntu.com/questions/64387/cannot-successfully-source-bashrc-from-a-shell-script
else
	echo "GOPATH exists already!"
exit 0
fi
