# Joaquin

A tool to save an image url to b2 cloud storage in periodic intervals
Part of the Nighthawk project

## Introduction

This cli is a simple data collection tool that connects to a http server and copying an image to an b2 bucket named 
by the date it was copied on.

## Usage
Install with `go get github.com/DerTiedemann/joaquin`

```
Usage of joaquin:
      --bucket string       b2 bucket name
      --id string           b2 account id
      --interval duration   interval between image fetches (default 10s)
      --key string          b2 api key
      --url string          url to the image
      --version             prints version
```

## Building

Build the binary for your system with `make build`
To build and install run `make install`



