#!/bin/sh

git clone --depth=1 git@github.com:/zhcppy/go-walletconnect-bridge
cd go-walletconnect-bridge
go build -o ../walletconnect-bridge
rm -rf ../go-walletconnect-bridge
