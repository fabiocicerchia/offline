# Specifications of offline

What this code does, one atomic Gherkin scenario per behaviour, read from the code by code-atlas.
Tags: `@verified` a test asserts it · `@inferred` read from the code only · `@gap` an entry point no test covers · `@human` written or amended by a person, never rewritten.

## Developer tooling
Build and quality commands for people working on offline itself.

- **Contributor setup**: [Discovering the available make targets](developer-tooling/contributor-setup/make-help.feature) (1)

## Installation and distribution
Ways to get offline onto a machine or into an image.

- **Container image**: [Running offline from the container image](installation/container-image/container-image.feature) (2)
- **Local install and uninstall**: [Installing and uninstalling offline locally](installation/local-install/local-install.feature) (3)
- **One-line installer**: [Installing offline with the one-line installer](installation/one-line-installer/one-line-installer.feature) (3)

## Network isolation
The guarantees that keep a wrapped program from reaching any network.

- **Isolated namespaces**: [Running a program in its own isolated namespaces](network-isolation/namespaces/isolated-namespaces.feature) (4)
- **Network call blocking**: [Failing network calls immediately](network-isolation/network-call-blocking/network-call-blocking.feature) (3)
- **Privilege dropping**: [Stopping privilege escalation inside the sandbox](network-isolation/privilege-dropping/privilege-dropping-offline.feature) (1), [Stripping privileges from the sandboxed program](network-isolation/privilege-dropping/privilege-dropping.feature) (2)
- **Socket family blocking**: [Refusing network sockets](network-isolation/socket-blocking/socket-family-blocking.feature) (3)

## Running a program
How a user launches a command under offline and what they get back.

- **Command line and usage**: [Invoking offline from the command line](program-execution/command-line/command-line.feature) (2)
- **Exit status passthrough**: [Passing the wrapped program's exit status back](program-execution/exit-status/exit-status-passthrough.feature) (3)
- **Terminal input and output**: [Connecting the wrapped program to the terminal](program-execution/terminal-io/terminal-io.feature) (1)

## Sandbox options
The settings a user can change to loosen or watch the isolation.

- **Blocked call logging**: [Reporting blocked network calls](sandbox-options/blocked-call-logging/blocked-call-logging.feature) (1)
- **Loopback access**: [Allowing local loopback traffic](sandbox-options/loopback-access/loopback-access-seccomp.feature) (1), [Keeping loopback reachable](sandbox-options/loopback-access/loopback-access.feature) (2)
- **Policy handoff**: [Carrying the chosen options into the sandbox](sandbox-options/policy-handoff/policy-handoff.feature) (2)
