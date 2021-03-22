There is no need for nasty quoting. Since quoting is for bash to pass arguments wholesale, you don't
need to return things with quotes. There is no bash waiting to remove them at the other end.

# Troubleshooting
Running your own shell-scripts.

Fatal error Error: fork/exec /home/jonasl/bin/local-auth-headers.sh: exec format error running command: /home/jonasl/bin/local-auth-headers.sh.

Remember to put a #!/bin/sh first in your scripts, otherwize exec does not know how to run the binary.