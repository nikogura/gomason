# gomason

[![Circle CI](https://circleci.com/gh/nikogura/gomason.svg?style=shield)](https://circleci.com/gh/nikogura/gomason)

[![Go Report Card](https://goreportcard.com/badge/github.com/nikogura/gomason)](https://goreportcard.com/report/github.com/nikogura/gomason)

Tool for testing, building, signing and publishing Go binaries in a clean Go workspace.  Think of it as an on premesis CI/CD system.

You could do this via a CI/CD System and an artifact repository of some flavor.  But wiring that up properly takes time, experience, and tends to be very specific to your particular system and repository.  

Gomason attempts to abstract all of that.  Any system must be able to handle the following:

1. Running tests and reporting on results

2. Building binaries for the target OS/Arch.

3. Personally signing the binaries thus built.  Artifact repos can sign things, it's true, but that just means you can verify that binary X was put in repo Y.  It's one step removed from actual authorship.  The idea is the actual author of a binary will sign their work- by whatever supported means you choose, and then you can upload it to the repo of your choice for distribution with the world.

4. Actually publish those binaries and their signatures to the artifact repo of your choice.

None of this is exactly rocket science, but the I have done it enough times, in enough different ways, that it was finally time to say 'enough' and be done with it.  

Gomason comes from an experience I had where management was so astounding **anti-testing** that I needed to come up with a way to do clean-room CI testing quickly, easily and transparently, but also fly under the radar.  They didn't need to know I was 'wasting time' testing my work. *(yeah, I couldn't believe it either.  Blew my mind.)* 

What started out as a sort of subversive method of continuing to test my own code expanded after we parted ways.  See, signing binaries and uploading the various bits to a repo isn't exactly rocket science, but it's also *dreadfully boring* once you've done it a few times.  I figured DRY, so I made a 'one and done' means for doing so.

CI systems like Artifactory Pro can sign binaries, it's true, but they don't really have provenance on who the author was.  The bits arrived there, presumably after authentication (but that depends on the config), but you don't really know who did the signing.

Enter gomason, which can do the building and signing locally with personal keys and then upload.  Presumably you'd also require authentication on upload, but now you've actually established 2 things- someone with credentials has uploaded this, *and* they've personally signed what they uploaded.  Whether you trust that signature is up to you, but we've provided an easy means to extend what a traditional CI system can do.

Gomason uses ```gox``` the Go cross compiler  to do it's compiling.  It builds whatever versions you like, but they need to be specified in the metadata.json file detailed below in gox-like format.

Code is downloaded via ```go get```.  If you have your VCS configured so that you can do that without authentication, then everything will *just work*.


## Installation

    go get github.com/nikogura/gomason
    
## Project Config

Projects are configured by the file ```metadata.json``` in the root of the project being tested/built/published by gomason.  This file is intended to be checked into the project and contains information required for gomason to function.  See below for examples and [Project Config Reference](#project-config-reference) for full details.

Some information in ```metadata.json```, such as signing info can be overwritten by the [User Config](#user-config) detailed below.

## User Config

User configuration is accomplished by the file ```~/.gomason```.  This is an *ini* formatted file that contains user level information such as the identity of the signer for use when signing binaries.

An example ```~/.gomason```:

    [user]
        email = nik.ogura@gmail.com

    [signing]
        program = gpg
        
This config would use the gpg program to sign binaries with the author's private key.  Obviously a key for the listed user must exist within gpg's keychain for this to function.

User config, if set, overrides any project config.

See [User Config Reference](#user-config-reference) for more details.
    
## Usage

* NOTE: you must have a ```metadata.json``` in your project as described in [Project Config Reference](#project-config-reference) below.

### Testing

Example Minimum Config:

    {
      "package": "github.com/nikogura/gomason",
      "version": "0.1.0",
      "description": "A tool for building and testing your project in a clean GOPATH."
    }

Run:

    gomason test
    
### Building

Example Minimum Config:

    {
      "package": "github.com/nikogura/gomason",
      "version": "0.1.0",
      "description": "A tool for building and testing your project in a clean GOPATH.",
      "building": {
          "targets": [
            "darwin/amd64",
            "linux/amd64"
          ]
      }
    }
    
## Project Config Reference

Gomason depends on a metadata file imaginatively named 'metadata.json'.  It's expected to be in the root of the repo.

The metadata.json contains such information as the version. (Yes, I'm old fashioned that way.  I like human readable version numbers.)

Example:

        {
          "version": "0.1.0",
          "package": "github.com/nikogura/gomason",
          "description": "A tool for building and testing your project in a clean GOPATH.",
          "building": {
              "targets": [
                "darwin/amd64",
                "linux/amd64"
              ]
          }
        }

### Version

Semantic version string of your package.  I realize go's github dependency mechanism provides commit-level granularity, but honestly?  Is that really useful?  

When's the last time you looked at a commit hash and derived any meaning around how much this version has changed from the last one you depended on?  I'm a fan of the idea that the major/minor/patch contract of semantic versioning can help you estimate, at a glance, how much of a change that upgrade you're pondering will be.

Sure, it needs to be tested.  (Trust but verify, right?)  But it's really nice to be able to have that estimate in a glance before you devote resources to the upgrade, even if it's just a quick estimate in your head.

### Package

The name of the Go package as used by 'go get' or 'govendor'.  Used to actually check out the code in the clean build environment.


### Description

A nice, human readable description for your module, cos that's really nice.  Having it in a parsable location as a simple string is also useful for other things, as you can probably imagine.

### Building

Information specifically for building the project.

#### Targets

This is used to determine which OSes and architectures to compile for. It's gotta be Gox's way of expressing the version and arch (os/arch), as the strings will simply be passed along to gox to build your toys.

For example, the following will build 64 bit binaries for MacOS and Linux:

    "targets": [
          "darwin/amd64",
          "linux/amd64"
    ]
    
    
### Signing

Information related to signing.

#### Program

Defaults to 'gpg'.  Others such as keybase.io will be added depending on time and user interest.

#### Email

The email of the entity (generally a person) who's doing the signing.  This entity, and their attendant keys must be available to the signing program.  

For instance, with the default 'gpg' program, gomason merely calls ```gpg -bau <email> <file>``` on the binaries.  If gpg doesn't already have a key registered for the email, an error will occur.

### Publishing

Information related to publishing.

## User Config Reference

Per-user config.  Primarily used to set per-user information that would not make sense to have in the project config.  

Which identity and key to use for signing is a good example.  While you can set and distribute a shared key for *everyone* to use, it's a better practice to have each publisher use their own key.  

The user config file gives you a place to do this.  You can, however set a group shared signing entity in ```metadata.json``` if you like. 

### User

The user using gomason.

#### Email

The email address of the person using gomason and signing binaries. 

example:

    [user]
        email = nik.ogura@gmail.com
        
### Signing

User specific configuration information related to signing.  Supported configuration keys:

#### Program

The program used to sign your binaries.  Set here it overrides any setting in ```metadata.json```

example:

    [signing]
        program = gpg
        
 

