## Lup - loopy command execution

Lup expands ranges and groups of terms in shell commands similarly to if you were using nested for loops.

`lup virsh @destroy,start@ @dev,test@_@1..3@`

Is functionally equivalent to:
```
for action in destroy start
do
  for environment in dev test
  do
    for i in 1 2 3
    do
      virsh $action $environment\_$i
    done
  done
done
```

Or simply:
```
virsh destroy dev_1
virsh destroy dev_2
virsh destroy dev_3
virsh destroy test_1
virsh destroy test_2
virsh destroy test_3
virsh start dev_1
virsh start dev_2
virsh start dev_3
virsh start test_1
virsh start test_2
virsh start test_3
```

Each command is run in sequence. In the event any command fails, lup will continue to trigger the remaining commands and will send 1 as its return code. Only if all commands run successfully will lup return 0.

## Trying it out

You can play around with lup in docker by running the following:

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

## Installing

Grab a release from the releases page, extract the binary, copy it to somewhere in your path (/usr/local/bin or /usr/bin are probably good choices) and set it to executable.

Something like this will do it if you have write access to /usr/local/bin

Mac: `curl -sL https://github.com/udkyo/lup/releases/download/v0.2.0/lup_0.2.0_darwin_amd64.tar.gz | tar xz lup && chmod +x lup ; mv lup /usr/local/bin`

Linux: `curl -sL https://github.com/udkyo/lup/releases/download/v0.2.0/lup_0.2.0_linux_amd64.tar.gz | tar xz lup && chmod +x lup ; mv lup /usr/local/bin`

## Compiling

Clone this repo and run `go build` then copy the generated executable into a pathed directory and set it to executable.

## Usage

### Dry Run

You can trigger a dry run by specifying -t as a flag (this must come immediately after "lup" on the command line - the rest of the line is treated as the command to be processed) this will echo the commands which lup intends to run, without actually triggering them.

Lup isn't very polished yet, so a dry run first is always a good idea.

### Spaces in terms

If you have spaces in any of your terms, you must encapsulate the group in either single or double quotes:

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

You can prevent terms from being used in a command by hiding them with -: e.g. `lup @-:0..10@ echo "Iteration @1@"` will echo the iteration 10 times, note you can still refer to these by index in a later backref. This can be helpful if you need to change the order commands run in. This can only be applied at the start of a group and will hide results for each term contained therein.

### Ranges

Ranges are available, but they must be the only thing contained within that group. They can count upwards or downwards, e.g. @1..100@ or @100..1@

### File Globbing

You can expand paths using standard globbing patterns using certain colon suffixed keywords. However the behaviour requires some explanation.

Loop for all dirs in /
`lup echo @dirs:/tmp/foo/*@`

Loop for all files in /
`lup echo @files:/tmp/foo/*@`

Loop for everything in /
`lup echo @all:/tmp/foo/*@`

What results is a list of file/dir names only, not the entire path - this makes it easier to manipulate names with backrefs later.

Consider the following:

`lup cp "/tmp/foo/@files:/tmp/foo/*@" "/tmp/bar/@dirs:/tmp/bar/*@/FILE_@1@"`

Two important things worth noting - dirs returned have no trailing slash, so we needed to add one after the @ group.

The other noteworthy things about this example is that by ensuring the path exists in the command prior to the @files:@ block, we filled the complete path. All the @files@ block captures is the filename so we can use it meaningfully elsewhere. If lup returned the entire path, copying and adding a prefix to a file like this would be more complicated.

However, the example above can be simplified further:

`lup -t cp "/tmp/foo/@files:*@" "/tmp/bar/@dirs:*@/FILE_@1@"`

When the files/dir/all blocks don't contain a path (i.e. they only contain the term followed by a colon) lup uses any path detected immediately prior to the block opening as the path processing should take place on, and all you need to provide to @files: is the pattern you would like to match. 

Or, to put it another way, these two things are more or less the same, but with one major difference which we'll discuss next:

`/tmp/foo/@files:*@`
`/tmp/foo/@files:/tmp/foo/*@`

Where lup's behavior varies is when pattern matching outside the @@s, lup references the *full file path* for each match. 

These will show the same results as we're only echoing the full path:

```
lup -t echo "/tmp/f*o/@files:*@"
lup -t echo "/tmp/foo/@files:*@"
```

But when trying to reference the filename via a backref it becomes obvious that the glob directive inheriting its path from outside the block is strictly full-paths only, while the other is file/node names alone.

```
lup -t echo "/tmp/f*o/@files:*@" @1@
lup -t echo "/tmp/foo/@files:*@" @1@
```

### Pipes and redirects

Lup won't straddle pipes or redirects, so if you are referencing terms on either side of those, it may be simplest to just pass the command as a string to a new shell as in the following example. 

`lup sh -c "echo @1..10@ > /tmp/@1@"`

### More on pipes

When piping a command's output to lup, that output will be captured and piped to each command lup generates and runs.

However (and this is important) when piping *from* lup, the output of each command lup runs will be merged and you'll probably end up having a pretty bad time. In general, you can encapsulate the whole command in a string and call a new shell with lup for each command it'll trigger:

```
lup sh -c "cat /opt/ssh/keys/training@1..10@.pub | ssh admin\\@train@1,2,3,4,5@.test 'cat >> ~training@1@/.ssh/authorized_keys'"
```

Or, you can just not use lup on the left hand side of your pipes (unless you really want all its output to be piped through in one go)

## Known issues

- Nesting isn't supported - if you run `lup nslookup @microsoft.@com,net,org@,google.com@` lup sees two groups - @microsoft.@ and @,google.com@ with the string com,net,org sandwiched in between
- ~- doesn't retrieve the previous working directory. I'm thinking tilde expansion should happen up front but that's not been the case in testing. Use a variable rather than relying on tilde expansion if you want previous working dir, $OLDPWD for example. On a related note, ~+ *does* work.
- at symbols make commands look cluttered - unfortunately all the more visually sensible choices with opening/closing pairs (parentheses, brackets, braces, chevrons) have built-in uses, so @ seems like the least idiotic character to use, however I'm open to suggestions
