## Lup - vertical command expansion

When used as a prefix for a command containing ampersand encapsulated, comma-separated groups of terms, lup processes from left to right, expanding terms and ranges as if they are nested for loops.

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

## Installing

Grab a release from the releases page, extract the binary, copy it to somewhere in your path (/usr/local/bin or /usr/bin are probably good choices) and set it to executable.

Something like this will do it if you have write access to /usr/local/bin

Mac: `curl -sL https://github.com/udkyo/lup/releases/download/v0.1.2/lup_0.1.2_darwin_amd64.tar.gz | tar xz lup && chmod +x lup ; mv lup /usr/local/bin`

Linux: `curl -sL https://github.com/udkyo/lup/releases/download/v0.1.2/lup_0.1.2_linux_amd64.tar.gz | tar xz lup && chmod +x lup ; mv lup /usr/local/bin`

## Compiling

Clone this repo and run `go build` then copy the generated executable into a pathed directory and set it to executable.

## Usage

### Dry Run

You can trigger a dry run by specifying -t as a flag, this will echo the commands which lup intends to run, without actually triggering them.

### Escaping control characters

Ampersands and commas within them are used as control characters, if you need to use these as normal characters, they should be escaped using slashes - note that slashes will need to be escaped themselves in certain situations. As a general rule of thumb you can use:

Double-slashes when enclosed in double quotes:
`lup echo "@Hello,Bonjour,Yo\\, wud up@ user\\@domain"`

Double-slashes when free standing :
`lup echo @Hello,Bonjour,Yo\\, wud up@ user\\@domain`

And single-slashes when enclosed in single quotes:
`lup echo '@Hello,Bonjour,Yo\, wud up@ user\@domain'`

### Escaping quotes

You'll find you need to escape quotes in certain circumstances - echo "hello world" comes through in os.Args as ["echo", "hello world"] so unless some kind soul can tell me something I've missed, I've got no visibility of the original quotes. Any spaced argument gets dropped into double-quotes by default and escaping happens automatically to the best of lup's ability, but there are situations where things break down.

`lup echo "this\"breaks"`

Is as good an example as any. Lup uses go-shellquote to split and recreate strings and in the example above, all it will be able to see at the end is "echo" whereas `lup echo "this\" works\"` (with a space) is fine. If you needed to have a single word with no spaces, and a double-quote in the middle, you could use `lup echo 'this\"works'` - using single quotes instead of doubles, while still escaping the double-quote (this is needed because lup plonks any spaced token into double quotes)

It's an edge case, but until I can work out a way to get around it it's worth mentioning - use -t first if you're getting fancy. 

### Referencing previous groups

To reuse a term (think: backrefs) you can use ampersands containing a single integer reference, these increment from 1, and the reference cannot come before the group it refers to.

Good:
`lup echo "@hello,goodbye@ @world,friend@ (@1@)"`

Bad:
`lup echo "@2@ @hello,goodbye@ @world,friend@"`

### Quick loops

You can loop commands without using the parsed term if the parsed term is the first thing in the command line and also a numeric range (e.g. `lup @0..10@ echo "Iteration @1@"` will echo the iteration 10 times, note I'm referring to the index with a backref here, but this isn't necessary)

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
- ampersands make commands look cluttered - unfortunately all the more visually sensible choices with opening/closing pairs (parentheses, brackets, braces, chevrons) have built-in uses, so @ seems like the least idiotic character to use, however I'm open to suggestions
- spaced arguments get plonked into double quotes before run, I don't have visibility of the original quotes in os.Args so I'm not sure how to fix this. The end result is if you have double quotes inside single quotes - `lup 'echo @hello,\"goodbye\"@ \"world\"'`, this is also true of freestanding single quotes with nothing between them, although single quotes inside double quotes are fine.
