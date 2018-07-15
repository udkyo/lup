package main

import (
	"fmt"
	"os"
)

func showHelp() {
	fmt.Println(`Usage: lup [OPTION] COMMANDLINE

Run multiple similar commands expanding at-symbol encapsulated, comma-separated lists similarly to nested for loops.

e.g:

  lup @rm,nano@ foo_@1,2@

Expands to and executes:

  rm foo_1
  rm foo_2
  nano foo_1
  nano foo_2

If @ symbols in a command or commas (within an @ group) are to be used as literals they should be escaped with backslashes, as should commas which are to be treated as literals within @ groups. When not enclosed in quotes, @s should be escaped with double-backslashes if intended as literals.

Iterating
---------
To iterate through a range of numbers, use @i..j@ where i and j are integer values. Lup will happily increment or decrement as required, for example.

  lup echo "@9..0@ @0..9@"

Hiding
------
To "hide" a group, you can prefix its contents with -: the following will echo iterate through the hidden block echoing "Hello" 5 times, but otherwise do nothing with its values

  lup @-:1..5@ echo "Hello"

Backrefs
--------
You can reference previous blocks in a command by including a standalone integer reference to it in an @ block. We can rework the previous example to echo the contents of the hidden block after the word "Hello" (note: backref values begin at 1 and are a copy of the specific value used in that group on any given line, they are not iterated through as independent loops)

  lup @-:1..5@ echo "Hello @1@"

Reading Files
-------------

Text files can be used in @ blocks and injected line by line. For example, given a file containing a list of servers:

  Fry
  Leela
  Bender

  We can perform an action on each line by using @lines:...@

  $ lup -t echo @lines:/tmp/foo/servers.txt@
  echo Fry
  echo Leela
  echo Bender

Filesystem
----------
A number of directives are available for iterating through files/directories/everything in a specific path. These are 'files,' 'dirs,' and 'all' respectively.

When provided with a relative path, the relative path and file/dir names will be used, e.g.:

  $ lup echo @files:foo/*@
  foo/bar.txt
  foo/baz.sh

When provided with an absolute path (whether explicit or expanded), only the file/directory name will be used:

  $ lup echo @files:$(pwd)/foo/*@
  bar.txt
  baz.sh

When a path is provided immediately before an @ group containing a files/dirs/all directive, the path is inherited:

  $ lup echo $(pwd)/foo/@files:*@
  /tmp/foo/bar.txt
  /tmp/foo/baz.sh

It is important to note that only the filename is being returned again - the full paths are shown because the rest of the path exists outside the @ group, we can prove this with a backref.

  $ lup echo "$(pwd)/foo/@files:*@ (@1@)"
  /tmp/foo/bar.txt (bar.txt)
  /tmp/foo/baz.sh (baz.sh)

If, however, a path is provided immediately prior to an @ group containing a files/dirs/all directive which contains a wildcard, the *entire path* will be returned.

  $ lup echo "$(pwd)/f*o/@files:*@ (@1@)"
  /tmp/foo/bar.txt (/tmp/foo/bar.txt)
  /tmp/foo/baz.sh (/tmp/foo/baz.sh)

More detail on usage is available at https://github.com/udkyo/lup

Options:

  -h, --help     Show this help message and exit
  -V, --version  Show version information and exit
  -t, --test     Show commands, but do not execute them`)
	os.Exit(0)
}
