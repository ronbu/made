[![Build Status](https://travis-ci.org/ronbu/made.png?branch=master)](https://travis-ci.org/ronbu/made)

Made is a simple reactive build tool.
If a file changes, every dependency of this file will be rebuilt.
This project is heavily inspired by [tup](http://gittup.org/tup/).

The build system needs a simple configuration file (Madefile) in the current
folder. It is composed of shell commands on every line.

Made watches all files enclosed in '<|' and '|>'. If the glob pattern
inside matches, the command on this line is executed.

	cp <|.hgignore|> .gitignore
	cp <|.gitignore|> .hgignore
	cp <|*.txt|> dir
	cp <|*.txt|> build/%b.html


The first line copies .hgignore to .gitignore if .hgignore is modified. The
second line does the same in reverse (probably not a good idea :-)). The third
line copies every changed *.txt file to the folder "dir". The last line copies
every *.txt file from the current folder to the folder "dir" and changes the
extension to .html, if a *.txt file has changed.

When made is executed it will run indefinitely and execute the above commands
on every matching file change event (newly added or modified).

# Installation

```bash
go get github.org/ronabu/made
```

If You have a properly configured [Go](https://golang.org) environment.

# TODO's and limitations

* Not stable yet
* Only OS X is currently supported.
* No concurrency
* Uses polling to find file modifications
* No string escaping in Madefile