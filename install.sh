#!/usr/bin/env bash

if [ $EUID -ne 0 ];
then
    echo "Please run this script as root"
    exit 1
fi

apt update && apt upgrade -y
apt install tor macchanger privoxy curl iptables-persistent psmisc -y

mkdir build/
go build -o build/torgo main.go

cp build/torgo /usr/bin/
cp -r torgo-iptables/ /usr/bin/

rm -r build/