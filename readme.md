# wzbug

This repository is reproduction of a bug found using wazero in the yoke project.

## What is it?

A go program compiled via the go toolchain that can be executed by a go program using wazero in version v1.6.0 hangs indefintily during instantiation with v1.8.1 and crashes completely with v1.7.3 

## Setup

The repository is setup to work as is. Simply run `./test.sh`.

It will compile the example program and try to execute it from Go using wazero for each wazero lib version.


## Wasi Package

The wasi package is a simplified version of the code from the yoke project used to execute wasm programs using wazero.
It is based on on the source code of `wazero run`.
