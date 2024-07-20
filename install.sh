#!/usr/bin/env bash

if [ $EUID -ne 0 ];
then
    echo "Please run this script as root"
    exit 1
fi

apt update && apt upgrade -y
apt install tor macchanger privoxy curl -y

mkdir build/
go build main.go -o build/