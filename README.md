## Lup - loopy command execution

Lup expands @ encapsulated blocks in shell commands similarly to if you were using nested for loops. 

Each command is run in sequence. In the event any command fails, lup will continue to trigger the remaining commands and will send 1 as its return code. Only if all commands run successfully will lup return 0.

[![Lup demo](https://raw.githubusercontent.com/udkyo/assets/master/lup2.gif)](https://asciinema.org/a/190728)

## Installing

Grab a release from the releases page, extract the binary, copy it to somewhere in your path (/usr/local/bin or /usr/bin are probably good choices) and set it to executable.

Something like this will do it if you have write access to /usr/local/bin

Mac: `curl -sL https://github.com/udkyo/lup/releases/download/v0.2.2/lup_0.2.2_darwin_amd64.tar.gz | tar xz lup && chmod +x lup ; mv lup /usr/local/bin`

Linux: `curl -sL https://github.com/udkyo/lup/releases/download/v0.2.2/lup_0.2.2_linux_amd64.tar.gz | tar xz lup && chmod +x lup ; mv lup /usr/local/bin`

If you want to just bring up a container with it to play around, just run:

```
docker run -it --rm alpine \
  sh -c \
  "apk add --no-cache curl \
   && curl -sL https://github.com/udkyo/lup/releases/download/v0.2.0/lup_0.2.0_linux_amd64.tar.gz \
   | tar xz lup \
   && chmod +x lup ; \
   mv lup /usr/local/bin && \
   sh"
```

## Usage

### Dry Run

You can trigger a dry run by specifying -t as a flag (this must come immediately after "lup" on the command line - the rest of the line is treated as the command to be processed) this will echo the commands which lup intends to run, without actually triggering them.

Lup isn't very polished yet, so a dry run first is always a good idea.

### Spaces in terms

If you have spaces in any of your terms, the group should be encapsulated in either single or double quotes:

`lup echo "@Hello,Well hello there@"`

### Escaping special characters

At symbols and commas (inside at symbol groups) are used as control characters, if you need to use these as normal characters, they should be escaped using slashes - note that slashes will need to be escaped with a single preceeding backslash

`lup echo '@Hello,Bonjour,Yo\, wud up@' user\@domain`

Whenever you escape commas or at symbols inside a group, you must also enclose the group in quotes.

### Referencing previous groups

To reuse a term (think: backrefs) you can use at symbols containing a single integer reference, these increment from 1, and the reference cannot come before the group it refers to.

Good:
`lup echo "@hello,goodbye@ @world,friend@ (@1@)"`

Bad:
`lup echo "@2@ @hello,goodbye@ @world,friend@"`

### Hiding terms

You can prevent terms from being used in a command by opening the block with -: e.g. `lup @-:0..10@ echo "Iteration @1@"` will echo the iteration 10 times, note you can still refer to these by index in a later backref. This can be helpful if you need to change the order commands run in.

Consider this somewhat ridiculous example:

`lup echo @1,2@ @3,4@ @5,6@`

The 8 combinations are echoed in sequence from "1 3 5" through to "2 4 6", with the first group being the outermost loop - the first 4 results will **start** with 1, and the last 4 results will **start** with 2

To switch this so that @1,2@ is the innermost loop, we could reverse the order the blocks appear in, and reference them in reverse order using backrefs:

`lup @-:5,6@ @-:3,4@ @-:1,2@ echo @3@ @2@ @1@`

With this change, @-:5,6@ is the outermost loop - the first 4 results will **end** with 5, and the final 4 will **end** with 6.

### Ranges

Numerical ranges are available, they can count upwards or downwards, e.g. @1..100@ or @100..1@

### File Globbing

You can expand paths using standard globbing patterns using colon suffixed keywords. However its behaviour varies based on context, it's important not to skim this section.

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
Notice directory names retrieved have no trailing slash.

### Pipes and redirects

Lup won't straddle pipes or redirects, so if you are referencing terms on either side of those, it is best to just pass the command as a string to a new shell as in the following example. 

`lup sh -c "echo @1..10@ > /tmp/@1@"`

### More on pipes

When piping a command's output to lup, that output will be captured and piped to each command lup generates and runs.

However (and this is important) when piping *from* lup, the output of each command lup runs will be merged and you'll probably end up having a pretty bad time. In general, you can encapsulate the whole command in a string and call a new shell with lup for each command it'll trigger:

```
lup sh -c "cat /opt/ssh/keys/training@1..10@.pub | ssh admin\\@train@1,2,3,4,5@.test 'cat >> ~training@1@/.ssh/authorized_keys'"
```

Or, you can just not use lup on the left hand side of your pipes (unless you really want all its output to be piped through in one go)

## Known issues

- Tilde completion immediately prior to a @ symbol is a no go. Instead you'll need to use full paths, $(pwd), $OLDPWD etc.
- Nesting isn't supported - if you run `lup nslookup @microsoft.@com,net,org@,google.com@` lup sees two groups - @microsoft.@ and @,google.com@ with the string com,net,org sandwiched in between
- at symbols make commands look cluttered - unfortunately all the more visually sensible choices with opening/closing pairs (parentheses, brackets, braces, chevrons) have built-in uses, so @ seems like the least idiotic character to use, however I'm open to suggestions
- lup triggers binaries, it doesn't operate on shell built-ins like set or export, so unfortunately you can't do things like `lup export @http,https@_proxy=http://foo/`
- command substitution happens up front before lup gets to work, bear that in mind if you're using $() or backticks inside a command that's being triggered by lup and considering putting @blocks@ in it
