# gomason

Tool for building Go binaries in a clean GOPATH.  

Why is this important?  Mainly to make sure you have all the dependencies properly vendored, and your tests run.

You don't care if your tests run?  That's a problem for the consumer of your code?

Go away choom.  Nothing more for us to talk about.  

If I have to explain why it's important to prove your code is working... nah.  You're supposed to grok that all on your own.

# Limitations

Right now, it's designed to work with git ssh repos with a url of the form 'git@github.com:<owner>/<repo>'
