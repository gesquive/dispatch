# dispatch

A self-hosted mail forwarding web API server.

This program was created to provide a json-based email API for static sites. You could use any transactional mailing service and pay them a monthly fee, or stand up your own dispatch.

## Installing

### Compile
This project has been tested with go1.7+. Just run `go get -u github.com/gesquive/dispatch` and the executable should be built for you automatically in your `$GOPATH`.

Optionally you can clone the repo and run `make install` to build and copy the executable to `/usr/local/bin/` with correct permissions.

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

### Targets
Targets are defined as follows:
```yaml
# auth-token should be a unique random string of characters
auth-token: f6uf9xvb@tze22O!KCZ7WExe
# emails will be sent from
from: dispatch@my-site.com
# emails will be sent too
to:
  - admin@my-site.com
  - personal@anywhere.com
```

Targets should be named with the `.yml` extension and be placed in the directory defined by the `--target-dir` flag. By default this is `/etc/dispatch/targets.d`.


### Environment Variables
Optionally, instead of using a config file you can specify config entries as environment variables. Use the prefix "DISPATCH_" in front of the uppercased variable name. For example, the config variable `smtp-server` would be the environment variable `DISPATCH_SMTP_SERVER`.

## Usage

```console
Run a webserver that provides an json api for emails

Usage:
  dispatch [flags]

Flags:
  -a, --address string         The IP address to bind the web server too (default "0.0.0.0")
      --check                  Check the config for errors and exit
      --config string          Path to a specific config file (default "./config.yml")
      --log-path string        Path to log files (default "/var/log/")
  -p, --port int               The port to bind the webserver too (default 8080)
  -r, --rate-limit string      The rate limit at which to send emails in the format 'inf|<num>/<duration>'.
                                 inf for infinite or 1/10s for 1 email per 10 seconds. (default "inf")
  -w, --smtp-password string   Authenticate the SMTP server with this password
  -o, --smtp-port uint32       The port to use for the SMTP server (default 25)
  -x, --smtp-server string     The SMTP server to send email through (default "localhost")
  -u, --smtp-username string   Authenticate the SMTP server with this user
      --target-dir string      Path to target configs (default "/etc/dispatch/targets.d")
  -v, --verbose                Print logs to stdout instead of file
      --version                Display the version number and exit
```

Optionally, a hidden debug flag is available in case you need additional output.
```console
Hidden Flags:
  -D, --debug                  Include debug statements in log output
```

### Service
This application was developed to run as a service behind a webserver such as nginx, apache, or caddy.

You can use upstart, init, runit or any other service manager to run the `dispatch` executable. Example scripts for systemd and upstart can be found in the `pkg/services` directory. A logrotate script can also be found in the `pkg/services` directory. All of the configs assume the user to run as is named `dispatch`, make sure to change this if needed.

## Documentation

This documentation can be found at github.com/gesquive/dispatch

## License

This package is made available under an MIT-style license. See LICENSE.

## Contributing

PRs are always welcome!
