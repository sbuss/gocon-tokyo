#!/bin/bash
set -x
go build .
sudo docker build . -t hello
sudo rm -rf /tmp/runsc
sudo docker run --runtime=runsc hello
