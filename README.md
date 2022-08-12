# honeypot

A simple telnet and ssh (for now) honeypot with a web interface.

![image](https://user-images.githubusercontent.com/5400940/184345780-80b7b429-c010-462f-81ad-d9f7dd75496e.png)


## Configuration

### Env variables

With default values

- `ADMIN_ADDR=localhost:7878` - where to listen for the web interface
- `HONEYPOT_DB=honeypot.db` - sqlite databasse location
- `TELNET_PORT=23` - poort to listen for telnet connections
- `IP2LOCATION_DB=IP2LOCATION-LITE-DB11.IPV6.BIN`
- `HONEYPOT_SSH_PORT=2222` - port to listen for ssh connections

## Building

To build the docker container

```sh
$ docker buildx build --platform linux/amd64 .
```

## Development

Supply IP2LOCATION-LITE-DB11.IPV6.BIN in the main directory of the project.

Then run

```sh
$ go run .
```

to run the backend

Then go to `honeypot-frontend` and run

```sh
$ npm i # install dependencies
$ npm run dev
```
