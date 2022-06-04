# honeypot

A simple telnet (for now) honeypot with a web interface.

## Configuration

### Env variables

With default values

- `ADMIN_ADDR=localhost:7878`
- `HONEYPOT_DB=honeypot.db`
- `TELNET_PORT=23`
- `IP2LOCATION_DB=IP2LOCATION-LITE-DB11.IPV6.BIN`
- `SHELL_PROVIDER_KEY=devkey`
- `HONEYPOT_SSH_PORT=2222`

## Building

To build the docker container

```sh
$ docker buildx build --platform linux/amd64 .
```
