# What?
Front-end for sending http requests via selectable backends (http, curl).

Has a global config with defaults (backend + any parameters (-sS))

Has a history list of the last N calls

Can save files from vim

With a flag includes the global config so it can be edited one-shot

Interpolates external commands ($uuidgen) and $env:ENV for environment variables.

Include a separate step for seeing the generated url (extra flag)

Sections are merged, with the local overriding the global (higher in file overrides lower in file)

# Syntax
[Variables]
ENV=test
GOAT=${TEST}
UUIDS=`
// Supports multiline
`

[Host]
http://localhost:8080/
http://scores.api-elb-${ENV}.internal.h11n.se/

[URL]

[Headers]
-H $TOKEN

[Method]
GET

[QPs]

[Body]

[Backend]
curl
-sS
--no-buffer

# Comment with global settings here, that can be uncommented
# [Backend]
# curl
# -sS
