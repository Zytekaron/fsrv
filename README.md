# fsrv

A personal and public file server which supports hosting files with
user-specific keys allowing them to read or modify files and directories.

## Key Features

### Permission Management

Files and directories can be locked to specific user tokens and set roles.
A file or directory can have permissions specifying (individually) that it
can only be read from or written to if the key is explicitly allowed or has
a role that is allowed by the file or directory.

Tokens and roles can also be explicitly denied access to a file which they
would otherwise have permission to access.

### Drop Requests

This file server facilitates owner-authenticated actions referred to as
Drop Requests, which allow for a way to request a file or files to be
placed into a directory by an otherwise unauthenticated user, without
providing information to that user about the result location of the files.

### Abuse Prevention

Multiple factors of abuse are prevented by setting a maximum depth of
subdirectories that can be created from the root of the file server, and
using rate limiting per key, as well as per ip for failed authentication.

# License
**fsrv** is licensed under the [MIT License](./LICENSE)
