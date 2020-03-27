# gomason

[![Current Release](https://img.shields.io/github/release/nikogura/gomason.svg)](https://img.shields.io/github/release/nikogura/gomason.svg)

[![Circle CI](https://circleci.com/gh/nikogura/gomason.svg?style=shield)](https://circleci.com/gh/nikogura/gomason)

[![Go Report Card](https://goreportcard.com/badge/github.com/nikogura/gomason)](https://goreportcard.com/report/github.com/nikogura/gomason)

[![Go Doc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](http://godoc.org/github.com/nikogura/gomason)

[![Coverage Status](https://codecov.io/gh/nikogura/gomason/branch/master/graph/badge.svg)](https://codecov.io/gh/nikogura/gomason)

[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go)  


Tool for testing, building, signing and publishing binaries.  Think of it as an on premesis CI/CD system- that also performs code signing and publishing of artifacts.

You could do this via a CI/CD System and an artifact repository of some flavor.  But wiring that up properly takes time, experience, and tends to be very specific to your particular system and repository.  

Gomason attempts to abstract all of that.  It will:

1. Run tests and report on results

2. Build binaries for the target OS/Arch and other files based on templates.

3. Sign the binaries and files thus built.

4. Publish the files, their signatures, and their checksums to the destination of your choice.

It does all of this based on config file called 'metadata.json' which you place in the root of your repository.

None of this is exactly rocket science, but I have done it enough times, in enough different ways, that it was finally time to say 'enough' and be done with it.  

Gomason comes from an experience I had where management was so astounding **anti-testing** that I needed to come up with a way to do clean-room CI testing quickly, easily and transparently, but also fly under the radar.  They didn't need to know I was 'wasting time' testing my work. *(yeah, I couldn't believe it either.  Blew my mind.)* 

What started out as a sort of subversive method of continuing to test my own code expanded after we parted ways.  See, signing binaries and uploading the various bits to a repo isn't exactly rocket science, but it's also *dreadfully boring* once you've done it a few times.  I figured DRY, so I made a 'one and done' means for doing so.

CI systems like Artifactory Pro can sign binaries, but they don't really have provenance on who the author was.  The bits arrived there, presumably after authentication (but that depends on the config), but you don't really know who did the signing.

Enter gomason, which can do the building and signing locally with personal keys and then upload.  Presumably you'd also require authentication on upload, but now you've actually established 2 things- someone with credentials has uploaded this, *and* they've personally signed what they uploaded.  Whether you trust that signature is up to you, but we've provided an easy means to extend what a traditional CI system can do.

Gomason uses ```gox``` the Go cross compiler  to do it's compiling.  It builds whatever versions you like, but they need to be specified in the metadata.json file detailed below in gox-like format.

Code is downloaded via ```go get```.  If you have your VCS configured so that you can do that without authentication, then everything will *just work*.

Signing is currently done via GPG.  I intend to support other signing methods such as Keybase.io, but at the moment, gpg is all you get.  If your signing keys are in gpg, and you have the gpg-agent running, it should *just work*.

## Language Suppport

At present, `gomason` only supports golang, but it turns out automating the whole test/build/sign/publish steps and making the UX painless is a really attractive thing, and it would be nice to have when working in other languages.

Basically, the authors have gotten so used to the painless UX that when we have to do things in other languages, we found ourselves missing `gomason`.  

Necessity being the mother of invention, stay tuned...

## Installation

    go get github.com/nikogura/gomason
    
## Usage

Test the master branch in a clean GOPATH return success/failure:

    gomason test
    
Test the master branch and see what's going on behind the scenes:

    gomason test -v
    
Test another branch verbosely:

    gomason test -v -b <branch name>
    
    
Publish the master branch after building:

    gomason publish -v
    
    
Build and publish a branch without testing (I know, I know, don't test?!!?!)  

This can be occasionally useful for publishing 3rd party tools internally when you need to make internal tweaks to support your use case.

Sometimes you don't have the wherewithal to build and maintain a full test suite for a 3rd party tool.  

    gomason publish -vs -b <branch name>
    
Other options can be found by running:

    gomason help
    
## Project Config

Projects are configured by the file ```metadata.json``` in the root of the project being tested/built/published by gomason.  This file is intended to be checked into the project and contains information required for gomason to function.  See below for examples and [Project Config Reference](#project-config-reference) for full details.

As of v2.6.1, `gomason` supports S3 urls of the 'virtual host' variety.  (i.e. `https://<bucket>.s3.<region>.amazonaws.com/<key>`).  It's assumed AWS credentials are configured in the environment used to run `gomason`.   If other AWS programs and tools work, `gomason` should too - so long as you have permission to write to the configured bucket(s).

Some information in ```metadata.json```, such as signing info can be overwritten by the [User Config](#user-config) detailed below.

Example metadata.json:

    {
      "version": "1.0.0",
      "package": "github.com/nikogura/gomason",
      "description": "A tool for testing, building, signing, and publishing your project from a clean workspace.",
      "repository": "http://localhost:8081/artifactory/generic-local",
      "tool-repository": "http://localhost:8081/artifactory/generic-local-tools",
      "insecure_get": false,
      "language": "golang",
      "building": {
        "prepcommands": [
          "go get k8s.io/client-go/...",
          "cd ${GOPATH}/src/k8s.io/client-go && git checkout v10.0.0",
          "cd ${GOPATH}/src/k8s.io/client-go && godep restore ./..."
        ],
        "targets": [
          {
            "name": "darwin/amd64",
            "cgo": true,
            "flags": {
              "CC": "o64-gcc",
              "CXX": "o64-g++"
             }
          {
            "name": "linux/amd64"
          }
        ]
      },
      "signing": {
        "program": "gpg",
        "email": "gomason-tester@foo.com"
      },
      "publishing": {
        "targets": [
          {
            "src": "gomason_darwin_amd64",
            "dst": "{{.Repository}}/gomason/{{.Version}}/darwin/amd64/gomason",
            "sig": true,
            "checksums": false
          },
          {
            "src": "gomason_linux_amd64",
            "dst": "{{.Repository}}/gomason/{{.Version}}/linux/amd64/gomason",
            "sig": true,
            "checksums": false
          }
        ]
      }
    }

## User Config

User configuration is accomplished by the file ```~/.gomason```.  This is an *ini* formatted file that contains user level information such as the identity of the signer for use when signing binaries.

An example ```~/.gomason```:

    [user]
        email = nik.ogura@gmail.com
        username = nikogura
        passwordfunc = lpass show --notes gomason-test

    [signing]
        program = gpg
        
This config would use the gpg program to sign binaries with the author's private key.  Obviously a key for the listed user must exist within gpg's keychain for this to function.  

This example also uses the LastPass cli to get the publishing password.  Neat huh?

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
            {
              "name": "darwin/amd64",
            {
              "name": "linux/amd64"
            }
          ]
      }
    }
    
Run:

    gomason build
    
The binaries will be moved into the current working directory.
    
### Signing

Example Config (Shared Key Signing):

    {
      "package": "github.com/nikogura/gomason",
      "version": "0.1.0",
      "description": "A tool for building and testing your project in a clean GOPATH.",
      "building": {
          "targets": [
            "darwin/amd64",
            "linux/amd64"
          ]
      },
      "signing": {
        "program": "gpg",
        "email": "gomason-tester@foo.com"
      },
    }
    
Run:

    gomason sign
    
The binaries and their signatures will be dumped into the current working directory.

Example Config (Personal Key Signing):

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
    
```~/.gomason```:


    [user]
        email = nik.ogura@gmail.com
        
    [signing]
        program = gpg
        
Run:

    gomason sign
    
The binaries and their signatures will be dumped into the current working directory.

### Publishing

Example Config (Personal Key Signing, Personal Credentials):

    {
      "package": "github.com/nikogura/gomason",
      "version": "0.1.0",
      "description": "A tool for building and testing your project in a clean GOPATH.",
      "repository": "http://localhost:8081/artifactory/generic-local",
      "building": {
          "targets": [
            {
              "name": "darwin/amd64",
            {
              "name": "linux/amd64"
            }
          ]
      },
      "publishing": {
        "targets": [
            {
                "src": "gomason_darwin_amd64",
                "dst": "{{.Repository}}/gomason/{{.Version}}/darwin/amd64/gomason",
                "sig": true,
                "checksums": false
            },
            {
                "src": "gomason_linux_amd64",
                "dst": "{{.Repository}}/gomason/{{.Version}}/linux/amd64/gomason",
                "sig": true,
                "checksums": false
            }
        ]
      }
    }
    
```~/.gomason```:


    [user]
        email = nik.ogura@gmail.com
        username = nikogura
        password = $ecretY0uNoR3ad!
        
    [signing]
        program = gpg
        
Run:

    gomason publish

Example Config (Shared Key Signing, Shared Credentials):

    {
      "package": "github.com/nikogura/gomason",
      "version": "0.1.0",
      "description": "A tool for building and testing your project in a clean GOPATH.",
      "repository": "http://localhost:8081/artifactory/generic-local",
      "building": {
          "targets": [
            {
              "name": "darwin/amd64",
            {
              "name": "linux/amd64"
            }
          ]
      },
      "signing": {
        "program": "gpg",
        "email": "gomason-tester@foo.com"
      },
      "publishing": {
        "targets": [
            {
                "src": "gomason_darwin_amd64",
                "dst": "{{.Repository}}/gomason/{{.Version}}/darwin/amd64/gomason",
                "sig": true,
                "checksums": false
            },
            {
                "src": "gomason_linux_amd64",
                "dst": "{{.Repository}}/gomason/{{.Version}}/linux/amd64/gomason",
                "sig": true,
                "checksums": false
            }
        ],
        "username": "nikogura",
        "password": "$ecretY0uNoR3ad!"
      }
    }
        
Run:

    gomason publish
    
### Publishing without Signing.
    
Occasionally, it might be useful to test and publish, but not sign.  Internal use for instance, where you don't really have a web of trust set up.

In that case, set "skip-signing": true in the publishing section and gomason will publish without bothering with the signatures.

Example:

    {
      "package": "github.com/nikogura/gomason",
      "version": "0.1.0",
      "description": "A tool for building and testing your project in a clean GOPATH.",
      "repository": "http://localhost:8081/artifactory/generic-local",
      "building": {
          "targets": [
            {
              "name": "darwin/amd64",
            {
              "name": "linux/amd64"
            }
          ]
      },
      "publishing": {
        "skip-signing": true,
        "targets": [
            {
                "src": "gomason_darwin_amd64",
                "dst": "{{.Repository}}/gomason/{{.Version}}/darwin/amd64/gomason",
                "sig": true,
                "checksums": false
            },
            {
                "src": "gomason_linux_amd64",
                "dst": "{{.Repository}}/gomason/{{.Version}}/linux/amd64/gomason",
                "sig": true,
                "checksums": false
            }
        ]
      }
    }

---
    
## Project Config Reference

Gomason depends on a metadata file imaginatively named 'metadata.json'.  It's expected to be in the root of the repo.

The metadata.json contains such information as the version. (Yes, I'm old fashioned that way.  I like human readable version numbers.)

Example:

    {
       "version": "0.1.0",
       "package": "github.com/nikogura/gomason",
       "description": "A tool for building and testing your project in a clean GOPATH.",
       "repository": "http://localhost:8081/artifactory/generic-local",
       "tool-repository": "http://localhost:8081/artifactory/generic-local-tools",
       "language": "golang",
       "building": {
         "prepcommands": [
              "go get k8s.io/client-go/...",
              "cd ${GOPATH}/src/k8s.io/client-go && git checkout v10.0.0",
              "cd ${GOPATH}/src/k8s.io/client-go && godep restore ./..."
         ],
         "targets": [
           {
                "name": "darwin/amd64",
           },
           {
                "name": "linux/amd64"
           }
         ]
       }
    }

### Version

Semantic version string of your package.  I realize go's github dependency mechanism provides commit-level granularity, but honestly?  Is that really useful?  

When's the last time you looked at a commit hash and derived any meaning around how much this version has changed from the last one you depended on?  I'm a fan of the idea that the major/minor/patch contract of semantic versioning can help you estimate, at a glance, how much of a change that upgrade you're pondering will be.

Sure, it needs to be tested.  (Trust but verify, right?)  But it's really nice to be able to have that estimate in a glance before you devote resources to the upgrade, even if it's just a quick estimate in your head.

### Package

The name of the Go package as used by 'go get'.  Used to actually check out the code in the clean build environment.


### Description

A nice, human readable description for your module, cos that's really nice.  Having it in a parsable location as a simple string is also useful for other things, as you can probably imagine.

### Repository

The url of the repository to which you're planning to publish your binaries.

### Tool-Repository

The url of a secondary repository to which you're planning to publish your binaries.  Primarily intended for use by (dbt)[https://github.com/nikogura/dbt].  

### Insecure_Get

Sometimes you've got a code repo that has a self signed cert.  Set this to true, and it'll pass ```-insecure``` to ```go get``` and ```govendor sync``` so you can still run- even if your internal repo has a self signed cert on it.

### Language

Optional at this point, and the only legal value is `golang`.  We plan to support other languages that can compile to single binaries in the future.

### Building

Information specifically for building the project.

#### Prepcommands

These are a list of bash commands that will be run prior to running any command that acutally uses your code, such as `go test'.

These commands are run one at a time, in a bash shell via `bash -c "<command"`.  This can be dangerous.  Use it with care. Obviously, bash has to exist on the system for things to work.  

This is primarily intended for situations like the Kubernetes Golang client, which needs special setup commands to pre-configure the dependencies in the GOPATH before actually testing and building your code.

#### Targets

This is used to determine which OSes and architectures to compile for. It's gotta be Gox's way of expressing the version and arch (os/arch), as the strings will simply be passed along to gox to build your toys.

Targets can take an optional 'cgo' flag to build with CGO, and a map of compiler flags that will be passed on to gox at build time.

Targets can also take an optional 'legacy' flag to build with GO111MODULE=off, for older projects that have not been converted yet.

This can be useful for cases where different targets require different options.

For example, the following will build 64 bit binaries for MacOS and Linux:

    "targets": [
      {
        "name": "darwin/amd64",
        "cgo": true,
        "flags": {
          "CC": "o64-gcc",
          "CXX": "o64-g++"
         }
      {
        "name": "linux/amd64"
      }
    ]
    
This of course, assumes you have gcc built able to cross-compile with something like https://github.com/tpoechtrager/osxcross.  The above works fine with MacOSX10.11 for the author.
    
#### Extras

Extra artifacts such as scripts and such you'd like built along side your go binaries.

Files are built from templates using the ```metadata.json``` as an information source.  Any field of the Metadata object created from ```metadata.json``` can be included in the template.  See golang's text/template documentation at [https://dlintw.github.io/gobyexample/public/text-template.html](https://dlintw.github.io/gobyexample/public/text-template.html) for examples.

Each 'extra' is a map with the following information:

* **template** String The template file to use.

* **filename** String The name of the file to write from the template

* **executable** Bool  Whether to make the written file executable.
    
### Signing

Information related to signing.

#### Program

Defaults to 'gpg'.  Others such as keybase.io will be added depending on time and user interest.

#### Email

The email of the entity (generally a person) who's doing the signing.  This entity, and their attendant keys must be available to the signing program.  

For instance, with the default 'gpg' program, gomason merely calls ```gpg -bau <email> <file>``` on the binaries.  If gpg doesn't already have a key registered for the email, an error will occur.

### Publishing

Information related to publishing.

#### Targets

Each target represents a file that will be uploaded.  Targets have the following attributes:

* **src** String. This is the file name as gomason would see it after running ```gox``` in the checked out code directory. 

* **dst** String. This is the upload path on the repository server.  Template fields of the form ```{{{.Field}}``` are supported.  The data being fed to the template is the Metadata object created from ```metadata.json```.  It's particularly useful for interpolating the *version* (```{{.Version}}```) and the *repository* ```{{.Repository}}``` into the upload path.

* **sig** Boolean.  Whether or not to upload the signature of the file you're publishing.  Generally you would want this to be true.

* **checksums** Boolean Whether or not to upload the checksum files for your published file.  Artifactory generates these files automatically, but if you're using something that supports a PUT, but can't generate the checksums, setting this to true will handle it for you.

#### Username

The username to use when authenticating to your artifact repository.  This can be set here, or in the per-user config.  Setting it in the per-user config is recommended.

#### Password

The password to use when authenticating to your artifact repository.  You can set it here (not recommended), or you can set it in the per-user config.

#### Usernamefunc

A shell function that will return the username to use.  Enables getting username info from a service.  Use carefully.  It's executing a command on your system.

#### Passwordfunc

A shell function that will return the password to use when publishing.  Enables getting the password from a service such as AWS Parameter store or Vault.

---

## User Config Reference

Per-user config.  Primarily used to set per-user information that would not make sense to have in the project config.  

Which identity and key to use for signing is a good example.  While you can set and distribute a shared key for *everyone* to use, it's a better practice to have each publisher use their own key.  

The user config file gives you a place to do this.  You can, however set a group shared signing entity in ```metadata.json``` if you like. 

### User

The user using gomason.

### Username

Username for your user.  This is used when publishing to an artifact repository.

example:

    [user]
        username = nikogura
        
#### Usernamefunc

A shell function that will return the username to use.  Enables getting username info from a service.  Use carefully.  It's executing a command on your system.  Probably not that useful, but supported for completenes sake.

example:

    [user]
        usernamefunc = curl -s http://url/of/config/service/where/we/store/the/username

### Password

Password for your user.  This is used when publishing to an artifact server.

example:
    
    [user]
        password = $ecretY0uNoR3ad!

#### Passwordfunc

A shell function that will return the password to use when publishing.  Really useful if you have a password manager with a cli such as LastPass.  You'll have to login separately though.  You won't be able to do it transparently via gomason... yet.  *(sometimes it takes a few tries to work out the magic)*

example:

    [user]
        passwordfunc = lpass show --notes gomason-test

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
        
 

