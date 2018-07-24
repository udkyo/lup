#!/bin/bash
set -e

echo "3 groups and a couple of backrefs:"
lup echo "  - @hello,goodbye@ @old,new@ @world,friend@ (@1@ @3@)"

echo "\nPipe:"
sh -c "echo '  - hello' | lup cat @-n,-e@"

echo "\nBackref hidden group"
lup @-:1..3@ echo "  - Iteration @1@"

echo "\nRange"
lup echo "  - @1..3@"

echo "\nVirsh"
lup echo "  - virsh @destroy,start@ @dev,test@_@1..3@"

echo "\nEscaped double quote inside single quotes"
lup echo '  - this\"works'

echo "\nEscaped double quote inside double quotes"
lup echo "  - this\"too"

echo "\nDoubles in singles"
lup echo '"hello world"'

echo "\nMisc quotes"
lup echo "  - "double quoted""
lup echo '  - 'single quoted''
lup echo "  - \"nested\""
lup echo "  - 'nested'"

echo "\nMatch files:"
lup echo "  - @files:/tmp/*@" "(@1@)"

echo "\nMatch dirs"
lup echo "  - @dirs:/tmp/*@" "(@1@)"

echo "\nMatch all"
lup echo "  - @all:/tmp/*@" "(@1@)"

echo "\nMatch files with prefix and wildcard"
lup echo "  - /tmp/*@files:*@" "(@1@)"

echo "\nMatch dirs with prefix and wildcard"
lup echo "  - /tmp/dirs/@dirs:*@" "(@1@)"

echo "\nMatch dirs with prefix and double wildcard"
lup echo "  - /tmp/d?rs/@dirs:*@" "(@1@)"

echo "\nMatch files with prefix and full path"
lup echo "  - /tmp/@files:*@" "(@1@)"

echo "\nMatch dirs with prefix and full path"
lup echo "  - /tmp/@dirs:*@" "(@1@)"

echo "\nLines in file"
echo "hello\nworld\nwhat's\nup?\nOh \"good" > /tmp/luptests/tests.txt
lup echo "  - @lines:/tmp/luptests/tests.txt@"
lup echo '  - @lines:/tmp/luptests/tests.txt@'
lup echo '  - nesting:"@lines:/tmp/luptests/tests.txt@"'
lup echo "  - nesting:'@lines:/tmp/luptests/tests.txt@'"
lup echo @lines:/tmp/luptests/tests.txt@

echo "\nEscaped at outside quotes"
lup echo "  - " root\\@lab-@0..2@.test

echo "\nEscaped at inside singles"
lup echo '  - root\@lab-@0..2@.test'

echo "\nEscaped at inside doubles"
lup echo "  - root\@lab-@0..2@.test"

echo "\nescaped comma inside, escaped ats in both (double quote):"
lup echo "  - @Hello,Bonjour,Yo\, \@wud up@ user\@domain"

echo "\nescaped comma inside, escaped ats in both (single quote):"
lup echo '  - @Hello,Bonjour,Yo\, \@wud up@ user\@domain'
