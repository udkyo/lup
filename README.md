## Lup - loopy command execution

[![Lup demo](https://raw.githubusercontent.com/udkyo/assets/master/lup2.gif)](https://asciinema.org/a/191914)

Lup runs whatever text you pass it on the command line as multiple separate commands, expanding the contents of @ encapsulated groups and directives vertically, in a similar manner to nested for loops.

e.g. `lup ping -c1 @google,amazon@.@com,net@` is similar to:

```
for site in google amazon
do
  for suffix in com net
  do 
    ping -c1 $site.$suffix
  done
done
```

Each command is run in sequence. In the event any command fails, lup will continue to trigger the remaining commands and will send 1 as its return code. Only if all commands run successfully will lup return 0.

## Table of Contents

  * [Lup - loopy command execution](#lup---loopy-command-execution)
  * [Table of Contents](#table-of-contents)
  * [Installing](#installing)
    * [Mac](#mac)
    * [Linux](#linux)
  * [Usage](#usage)
    * [Dry run](#dry-run)
    * [Escaping special characters](#escaping-special-characters)
    * [Ranges](#ranges)
    * [Backrefs](#backrefs)
    * [Hidden groups](#hidden-groups)
    * [File Globbing](#file-globbing)
    * [Reading a file](#reading-a-file)
    * [Pipes and redirects](#pipes-and-redirects)
    * [More on pipes](#more-on-pipes)
  * [Known issues](#known-issues)

## Installing

#### Mac

Brew users on mac can install with:

```
brew tap udkyo/lup
brew install lup
```

Or manually:

```
curl -sL https://github.com/udkyo/lup/releases/download/v0.4.0/lup_0.4.0_darwin_amd64.tar.gz \
  | tar xz lup \
  && chmod +x lup
mv lup /usr/local/bin
```

#### Linux

Download and extract the latest release, set the file "lup" to executable and move it to a pathed directory. e.g:

```
curl -sL https://github.com/udkyo/lup/releases/download/v0.4.0/lup_0.4.0_linux_amd64.tar.gz \
  | tar xz lup \
  && chmod +x lup ;
mv lup /usr/local/bin
```

## Usage

### Dry run

You can trigger a dry run by specifying -t as a flag, this will show the commands which lup intends to run without actually triggering them.

Note: lup's flags must always be the first things on the command line after the word 'lup' - everything else gets treated as the command lup should expand.

Another note: Doing a dry run first is always a good idea, at least until you're comfortable with how lup works.

### Escaping special characters

@ symbols anywhere in the command, and commas inside @ groups are used as control characters, if you need to use these as normal characters, they should be escaped using slashes:

`lup echo '@Hello,Bonjour,Yo\, wud up@' user\@domain`

When escaping outside quotes, you will need to use double-slashes

### Ranges

Numerical ranges are available, they can count upwards or downwards, e.g. `@1..100@` or `@100..1@`

### Backrefs

To reuse a term you can use @ groups containing a single integer reference, these increment from 1, and the reference cannot come before the group it refers to.

Good:

`lup echo "@hello,goodbye@ @world,friend@ (@1@)"`

Bad:

`lup echo "@2@ @hello,goodbye@ @world,friend@"`

### Hidden groups

You can prevent terms from being used in a command by opening the block with `-:` e.g. `lup @-:0..10@ echo "Iteration @1@"` will echo the iteration 10 times, note you can still refer to these by index in a later backref. This can be helpful if you need to change the order commands run in.

Consider the following:

```
$ lup echo @1,2@ @3,4@ @5,6@
1 3 5
1 3 6
1 4 5
1 4 6
2 3 5
2 3 6
2 4 5
2 4 6
```

Clearly, lup treats the first group encountered as the outermost loop, and the final one as the innermost.

To switch this so that @1,2@ is the innermost loop, we can use hiding, reverse the order, and display them in the correct order using backrefs:

```
$ lup @-:5,6@ @-:3,4@ @-:1,2@ echo @3@ @2@ @1@
1 3 5
2 3 5
1 4 5
2 4 5
1 3 6
2 3 6
1 4 6
2 4 6
```

### Reading a File

Text files can be used in @ blocks and injected line by line. For example, given a file containing a list of servers:

```
Fry
Leela
Bender
```

We can perform an action on each line by using @lines:...@

```
$ lup echo @lines:/tmp/foo/servers.txt@
Fry
Leela
Bender
```

### File Globbing

You can expand paths using standard globbing patterns with colon suffixed keywords. However its behaviour varies if a path immediately precedes the group.

Note: Paths outside the @ blocks are matched by simply walking through the command from left to right, checking for slashes and allowing paths to expand out from there, so *anything* that looks like a path which appears immediately before a block will currently be treated as one. e.g. `this/that@files:*@` will try to retrieve a list of files in the directory `/that`.

#### Basic directives:

`@dirs:/tmp/foo/*@` - All directories in /tmp/foo/

`@files:/tmp/foo/*@` - All files in /tmp/foo/

`@all:/tmp/foo/*@` - Everything in /tmp/foo/

Let's look more closely at the behavior of these directives by running some commands against a directory /tmp/lup, which has the subdirectories foo, bar and baz:

When matching via an absolute path, only the name of the folder is returned

```
$ lup echo "@dirs:/tmp/lup/*@"
bar
baz
foo
```

When referencing a relative path, the relative path is returned

```
$ cd /tmp
$ lup echo "@dirs:lup/*@"
lup/bar
lup/baz
lup/foo
```

If an absolute path precedes the block, matches in the block are relative to the external path, and only the relative path is returned (I use a backref here for clarity, as /tmp/lup/ remains in the output by virtue of existing outside the block)

```
$ lup echo "/tmp/lup/@dirs:*@ - returned @1@"
/tmp/lup/bar - returned bar
/tmp/lup/baz - returned baz
/tmp/lup/foo - returned foo
```

When an absolute path precedes the block and contains a pattern to be matched, the *absolute path* is returned:

```
lup echo "/tmp/l?p/@dirs:*@ - returned @1@"
/tmp/lup/bar - returned /tmp/lup/bar
/tmp/lup/baz - returned /tmp/lup/baz
/tmp/lup/foo - returned /tmp/lup/foo
```
Notice directory names retrieved never feature a trailing slash.

### Pipes and redirects

Lup won't straddle pipes or redirects, so if you are referencing terms on either side of those, it is best to just pass the command as a string to a new shell as in the following example. 

`lup sh -c "echo @1..10@ > /tmp/@1@"`

### More on pipes

When piping a command's output to lup, that output will be captured and piped to each command lup generates and runs.

However, when piping *from* lup, the output of each command lup runs will be merged and you'll probably end up having a pretty bad time. In general, you can encapsulate the whole command in a string and call a new shell with lup for each command it'll trigger:

```
lup sh -c "cat /opt/ssh/keys/training@1..10@.pub | ssh admin\@train@1,2,3,4,5@.test 'cat >> ~training@1@/.ssh/authorized_keys'"
```

Or, you can just not use lup on the left hand side of your pipes (unless you really want all its output to be piped through in one go)

## Known issues

- Tilde completion immediately prior to a @ symbol is a no go. Instead you'll need to use full paths, $(pwd), $OLDPWD etc.
- Nesting isn't supported - if you run `lup nslookup @microsoft.@com,net,org@,google.com@` lup sees two groups - @microsoft.@ and @,google.com@ with the string com,net,org sandwiched in between
- at symbols make commands look cluttered - unfortunately all the more visually sensible choices with opening/closing pairs (parentheses, brackets, braces, chevrons) have built-in uses, so @ seems like the least idiotic character to use, however I'm open to suggestions
- lup triggers binaries, it doesn't operate on shell built-ins like set or export, so unfortunately you can't directly do actions such `lup export http@,s@_proxy=http://foo/`, however you can circumvent this using builtin, e.g. `lup builtin export http@,s@_proxy="http://foo/"`
- command substitution happens up front before lup gets to work, bear that in mind if you're using $() or backticks inside a command that's being triggered by lup and considering putting @ blocks in it
