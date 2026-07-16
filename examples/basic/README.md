# Basic Example

What it shows: running a command inside `offline` and confirming it has no
network access.

## Run

From the repo root, after `make build`:

```sh
./offline curl https://example.com
```

Expected:

```
curl: (6) Could not resolve host: example.com
```

Try an offline shell:

```sh
./offline bash
```

Inside it, `ip addr` shows only `lo`, `ip route` shows nothing, and any
`curl`/`wget` fails to resolve a host.
