#!/bin/bash

RUN_NAME="test"

mkdir -p ./output/bin ./output/conf
go build -mod=vendor -a -o ./output/bin/${RUN_NAME}

./output/bin/test


