# gomason

[![Circle CI](https://circleci.com/gh/nikogura/gomason.svg?style=shield)](https://circleci.com/gh/nikogura/gomason)

[![Go Report Card](https://goreportcard.com/badge/github.com/nikogura/gomason)](https://goreportcard.com/report/github.com/nikogura/gomason)

Tool for building Go binaries in a clean GOPATH.  

Why is this important?  Mainly to make sure you have all the dependencies properly vendored, and your tests run.

You don't care if your tests run?  That's a problem for the consumer of your code?

Go away choom.  Nothing more for us to talk about.  

If I have to explain why it's important to prove your code is working... nah.  You're supposed to grok that all on your own.

Gomason uses gox to do it's compiling.  It builds whatever versions you like, but they need to be specified in the metadata.json file detailed below in gox-like format.

## Config

Gomason depends on a metadata file imaginatively named 'metadata.json'.  It's expected to be in the root of the repo.

The metadata.json contains such information as the version. (Yes, I'm old fashioned that way.  I like human readable version numbers.)

The 'buildtargets' section of the metadata.json is used to determine which OSes and architectures to compile for.  The 'package' section is the package spec as used by 'go get' and 'govendor'.

## Limitations

Right now, it's designed to work with git ssh repos with a url of the form 'git@github.com:(owner)/(repo)'
