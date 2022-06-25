
#!/bin/sh
ulimit -n 100000
export GOROOT=/usr/local/go
export GOPATH=~/kekops
export PATH=$GOPATH/bin:$GOROOT/bin:$PATH
