#!/bin/sh

rm -rf -- "/tmp/wasp"
mkdir -p "/tmp/wasp"
cd "/tmp/wasp"

mkdir -p a/a && mkdir -p a/b
echo "test1" > a/a/1.txt
echo "test2" > a/b/2.txt
touch a/empty.txt

