# dispatch

A self-hosted mail forwarding web API server.

### Why?
This program was created to provide an email json-based API for static sites.

## Installing

### Compile
This project requires go1.7+ to compile. Just run `go get -u github.com/gesquive/dispatch` and the executable should be built for you automatically in your `$GOPATH`.

Optionally you can run `make install` to build and copy the executable to `/usr/local/bin/` with correct permissions.

### Download
Alternately, you can download the latest release for your platform from [github](https://github.com/gesquive/dispatch/releases).

Once you have an executable, make sure to copy it somewhere on your path like `/usr/local/bin` or `C:/Program Files/`.
If on a \*nix/mac system, make sure to run `chmod +x /path/to/dispatch`.

## Configuration

### Precedence Order
The application looks for variables in the following order:
 - command line flag
 - environment variable
 - config file variable
 - default

So any variable specified on the command line would override values set in the environment or config file.

### Config File
The application looks for a configuration file at the following locations in order:
 - `./config.yml`
 - `~/.config/dispatch/config.yml`
 - `/etc/dispatch/config.yml`

Copy `pkg/config.example.yml` to one of these locations and populate the values with your own. Since the config contains a writable API token, make sure to set permissions on the config file appropriately so others cannot read it. A good suggestion is `chmod 600 /path/to/config.yml`.

If you are planning to run this app as a service, it is recommended that you place the config in `/etc/dispatch/config.yml`.

### Environment Variables
Optionally, instead of using a config file you can specify config entries as environment variables. Use the prefix "DISPATCH_" in front of the uppercased variable name. For example, the config variable `smtp-server` would be the environment variable `DISPATCH_SMTP_SERVER`.

## Usage

```console
This app runs a webserver that provides an api for email forwards

Usage:
  dispatch [flags]

Flags:
  -a, --address string         The IP address to bind the web server too (default "0.0.0.0")
      --check                  Check the config for errors and exit
      --config string          Path to a specific config file (default "./config.yaml")
      --log-path string        Path to log files (default "/var/log/")
  -p, --port int               The port to bind the webserver too (default 8080)
  -w, --smtp-password string   Authenticate the SMTP server with this password
  -o, --smtp-port value        The port to use for the SMTP server (default 25)
  -x, --smtp-server string     The SMTP server to send email through (default "localhost")
  -u, --smtp-username string   Authenticate the SMTP server with this user
  -v, --verbose                Print logs to stdout instead of file
      --version                Display the version number and exit
```

It is helpful to use the `--run-once` combined with the `--verbose` flags when first setting up to find any misconfigurations.

Optionally, a hidden debug flag is available in case you need additional output.
```console
Hidden Flags:
  -D, --debug                  Include debug statements in log output
```

### Service
By default, the process is setup to run as a service. Feel free to use upstart, init, runit or any other service manager to run the `dispatch` executable.

## Documentation

This documentation can be found at github.com/gesquive/dispatch

## License

This package is made available under an MIT-style license. See LICENSE.

## Contributing

PRs are always welcome!


<!-- TODO: Include some default upstart/init scripts -->
<!-- TODO: Include a logrotate script -->
<!-- TODO: Create a detailed service install script -->
