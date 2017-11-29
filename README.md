# gomason

[![Circle CI](https://circleci.com/gh/nikogura/gomason.svg?style=shield)](https://circleci.com/gh/nikogura/gomason)

[![Go Report Card](https://goreportcard.com/badge/github.com/nikogura/gomason)](https://goreportcard.com/report/github.com/nikogura/gomason)

Tool for building Go binaries in a clean GOPATH.  

Why is this important?  Mainly to make sure you have all the dependencies properly vendored, and your tests run.

You don't care if your tests run?  That's a problem for the consumer of your code? Go away choom.  Nothing more for us to talk about.  

If I have to explain why it's important to prove your code is working... nah.  You're supposed to understand that all on your own.

If you have access to a CI system, ```gomason``` isn't going to impress you- unless like me you like making sure all t's are crossed and all i's are dotted *before* you commit and push.  Pushing things that fail is harmless if you have a good branching model and you adhere to it, but it's still embarassing.

Gomason comes from an experience where management was so astounding **anti-testing** *(yeah, I couldn't believe it either.)* that I needed to come up with a way to do clean-room CI testing quickly, easily and transparently, but also fly under the weather.  They didn't need to know I was 'wasting time' testing my work.

Gomason uses ```gox``` the Go cross compiler  to do it's compiling.  It builds whatever versions you like, but they need to be specified in the metadata.json file detailed below in gox-like format.

## Config

Gomason depends on a metadata file imaginatively named 'metadata.json'.  It's expected to be in the root of the repo.

The metadata.json contains such information as the version. (Yes, I'm old fashioned that way.  I like human readable version numbers.)

Example:

        {
          "version": "0.1.0",
          "package": "github.com/nikogura/gomason",
          "description": "A tool for building and testing your project in a clean GOPATH.",
          "buildtargets": [
            "darwin/amd64",
            "linux/amd64"
          ]
        }

### Config Sections

#### Version

Semantic version string of your package.  I realize go's github dependency mechanism provides commit-level granularity, but honestly?  Is that really useful?  

When's the last time you looked at a commit hash and derived any meaning around how much this version has changed from the last one you depended on?  I'm a fan of the idea that the major/minor/patch contract of semantic versioning can help you estimate, at a glance, how much of a change that upgrade you're pondering will be.

Sure, it needs to be tested.  (Trust but verify, right?)  But it's really nice to be able to have that estimate in a glance before you devote resources to the upgrade, even if it's just a quick estimate in your head.

#### Package

The name of the Go package as used by 'go get' or 'govendor'.  Used to actually check out the code in the clean build environment.


#### Description

A nice, human readable description for your module, cos that's really nice.  Having it in a parsable location as a simple string is also useful for other things, as you can probably imagine.

#### Buildtargets

This is used to determine which OSes and architectures to compile for. It's gotta be Gox's way of expressing the version and arch (os/arch), as the strings will simply be passed along to gox to build your toys.

## Limitations

Right now, it's designed to work with git ssh repos with a url of the form 'git@github.com:(owner)/(repo)'
