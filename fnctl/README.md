# IronFunctions CLI

You can operate IronFunctions from the command line.

```ShellSession
$ fnctl apps                                       # list apps
myapp

$ fnctl apps create otherapp                       # create new app
otherapp created

$ fnctl apps
myapp
otherapp

$ fnctl routes myapp                               # list routes of an app
path	image
/hello	iron/hello

$ fnctl routes create otherapp /hello iron/hello   # create route
/hello created with iron/hello
```