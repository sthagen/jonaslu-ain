# Design
The file to parse is passed as the last argument (or can take several arguments?)

Once one is selected, we also look in ~/.ain for any global file
Those sections is inserted into a temp-file as comments

Then we fire up vim with the resulting file (read from /tty/stdin always)

Once done with editing:
We parse and validate it:

Create an 'AST' with one object per section. Each section is its own file (so its a nice pluggable system):

We have transforms that take the text and perform string-substitution

main.go
parse/editor.go
parse/parse.go
sections/host.go
sections/URL.go
transforms/variables.go <- Scan text for ${}
transforms/subshell.go <- Scan text for $()
backends/curl.go
backends/http.go
history/history.go <- Save a validated call into the history file
output/output.go <- May know about MIME-tags? Or let others format shit?

Each section can then contain interpolations or not:
$VAR is overridden in the [Variable] section. If not found there we try to fetch it from the environment
$() is a subshell command, it's run as such and the result is inserted

As input, expects the file to run as the last argument (xargs friendly).
If input is a pipe, read the file-name to run.

This enables us to chain fzf to find all .ain files as such:
find . -name *.ain | fzf | ain |

Make vim read from /dev/tty instead so we can be in a pipe

If not connected to a pipe, a switch can print out the resulting command before it's run (good for introspection)

Has a list of the last N calls (store them in ~/.ain/history) - if -he is given you can edit
the call (and it is saved as a new call). Could possibly show the resulting call as a short reminder of what in
the list.

# Main
Checks if we're connected to a pipe, reads the filename from stdin if so (test with fzf)
OR
Takes command line and zero to 1 files to call (separates call by a delimiter)

-g (include global parameters)
-h (history)
-e (edit - if history)
-w (write the resulting command line to file instead of running it - useful for generating cron lines or debugging)

calls editor/parse with the sent or fetched file - capture the result
saves the call into the history file
runs the call or writes it to disk

# parse/editor.go
Saves the read file to a temp-file and opens up vim connected via /dev/tty.
Captures the edited result.

# parse/parse.go
Discard any comments in the source.
Finds all headers and trims their whitespace

Checks all [XX] headers are valid (by finding their parser)
Has a big struct for all valid headers:

type struct yak {
  Host string
  URL string
  Body string
  Parameters []string
  Method string
}

Variables are captured outside the string as it's epmhemeral:
Variables []string

Lets the sections fill in parameters
Then passes all sections into the transformators (by passing string or []string).
Not all sections may be transformed.

Passes the subsection to each parser for parser.preProcess() error
The error signifies that the section was not valid.

Then passes each section into the transforms.

# backends/curl
Is passed the big struct and then assembles it as a command. Returns the command as a string.
Should error if the data cannot be assembled as a valid call.
Should warn if some combinations are invalid (such as )

# MVP
Have a ready file, simplest one: [Host] and [URL]
Pass it into vim (no /dev/tty)
Strip comments, validate headers
Pass it into [Host] and [URL]
Pass it into backend to assemble it
Run it

# MVP 2
Variable interpolation, add ${} substitution

# MVP 3
Subshell interpolation, add $() subshell running

# MVP 4
Add a global settings file that is silently included (and is overriden by local declarations)
