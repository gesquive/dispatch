# dispatch
[![Software License](https://img.shields.io/badge/License-MIT-orange.svg?style=flat-square)](https://github.com/gesquive/dispatch/blob/master/LICENSE)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](https://godoc.org/github.com/gesquive/dispatch)
[![Pipeline Status](https://img.shields.io/gitlab/pipeline/gesquive/dispatch?style=flat-square)](https://gitlab.com/gesquive/dispatch/pipelines)
[![Coverage Report](https://gitlab.com/gesquive/dispatch/badges/master/coverage.svg?style=flat-square)](https://gesquive.gitlab.io/dispatch/coverage.html)
[![Docker Pulls](https://img.shields.io/docker/pulls/gesquive/dispatch?style=flat-square)](https://hub.docker.com/r/gesquive/dispatch)


A self-hosted mail forwarding web API server.

This program was created to provide a json-based email API for static sites. You could use any transactional mailing service and pay them a monthly fee, or stand up your own dispatch.

## Installing

### Compile
This project has only been tested with go1.11+. To compile just run `go get -u github.com/gesquive/dispatch` and the executable should be built for you automatically in your `$GOPATH`. This project uses go mods, so you might need to set `GO111MODULE=on` in order for `go get` to complete properly.

Optionally you can clone the repo and run `make install` to build and copy the executable to `/usr/local/bin/` with correct permissions.

### Download
Alternately, you can download the latest release for your platform from [github](https://github.com/gesquive/dispatch/releases).

Once you have an executable, make sure to copy it somewhere on your path like `/usr/local/bin` or `C:/Program Files/`.
If on a \*nix/mac system, make sure to run `chmod +x /path/to/dispatch`.

### Docker
You can also run dispatch from the provided [Docker image](https://hub.docker.com/r/gesquive/dispatch) with the sample configuration file:

```shell
mkdir -p dispatch && cp pkg/config.example.yml dispatch/config.yml
docker run -d -p 2525:2525 -v $PWD/dispatch:/config dispatch:latest
```

To get the sample config working, you will need to configure the SMTP server and add target configs. 

For more details read the [Docker image documentation](https://hub.docker.com/r/gesquive/dispatch).

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
# name is the unique name for this target (default value is the target filename)
# name will show up in the subject of the email
name: example
# auth-token should be a unique random string of characters
auth-token: f6uf9xvb@tze22O!KCZ7WExe
# emails will be sent from
from: dispatch@my-site.com
# emails will be sent too
to:
  - admin@my-site.com
  - personal@anywhere.com

defaults:
  subject: "site message"
  tag: "no-tag"

```

Targets should be named with the `.yml` extension and be placed in the directory defined by the `--target-dir` flag. By default this is `/etc/dispatch/targets-enabled`.

#### Target Auth Tokens
Each target requires a unique Auth token so incoming messages can be routed to the correct target. Without a unique auth tokens, messages will be routed incorrectly.

#### Target Defaults
Any key-values specified under the `defaults` variable will be used as the default for all incoming requests. Any values specified in either the payload or header will overwrite these values

#### Request HTTP Headers
It is possible to specify a value for a request in the HTTP headers. Values specified in an HTTP header will always overwrite values specified through json. To specify a value through HTTP headers use the prefix `X-Dispatch-` with the variable name. For example, if you wanted to specify the `auth-token` through a HTTP header, simply post the json with the header `X-Dispatch-Auth-Token`

#### Request Precedence Order
Values for requests can be specified through an http request, request header or target default. The application takes values in the following order:
 - request header
 - request payload
 - target default

 So a variable value specified in an http request header will always override a value specified in the payload of an http request.

### Environment Variables
Optionally, instead of using a config file you can specify config entries as environment variables. Use the prefix `DISPATCH_` in front of the uppercased variable name. For example, the config variable `smtp-server` would be the environment variable `DISPATCH_SMTP_SERVER`.

### Service
This application was developed to run as a service behind a webserver such as nginx, apache, or caddy.

You can use upstart, init, runit or any other service manager to run the `dispatch` executable. Example scripts for systemd and upstart can be found in the `pkg/services` directory. A logrotate script can also be found in the `pkg/services` directory. All of the configs assume the user to run as is named `dispatch`, make sure to change this if needed.

## Usage

```console
Run a webserver that provides an json api for emails

Usage:
  dispatch [flags]

Flags:
  -a, --address string         The IP address to bind the web server too (default "0.0.0.0")
      --check                  Check the config for errors and exit
      --config string          Path to a specific config file (default "./config.yml")
  -l, --log-file string        Path to log file (default "/var/log/dispatch.log")
  -p, --port int               The port to bind the webserver too (default 2525)
  -r, --rate-limit string      The rate limit at which to send emails in the format 'inf|<num>/<duration>'. inf for infinite or 1/10s for 1 email per 10 seconds. (default "inf")
  -w, --smtp-password string   Authenticate the SMTP server with this password
  -o, --smtp-port uint32       The port to use for the SMTP server (default 25)
  -x, --smtp-server string     The SMTP server to send email through (default "localhost")
  -u, --smtp-username string   Authenticate the SMTP server with this user
  -t, --target-dir string      Path to target configs (default "/etc/dispatch/targets-enabled")
      --version                Display the version number and exit
```

Optionally, a hidden debug flag is available in case you need additional output.
```console
Hidden Flags:
  -D, --debug                  Include debug statements in log output
```

## Examples
To send an email using dispatch, simply send a JSON formatted POST request to the `/send` endpoint. The format is as follows:
```json
{
    "auth-token": "",
    "name": "",
    "email": "",
    "subject": "",
    "message": "",
}
```

`auth-token` is the only required field. If not provided in the json as `auth-token` it must be passed through the HTTP Header `X-Dispatch-Auth-Token`. dispatch also checks to see if the `email` field is a valid email address.

### Javascript example
```javascript
$(document).ready(function() {

    // process the form
    $('form').submit(function(e) {

        // get the form data
        var formData = {
            'name'          : $('input[name=name]').val(),
            'email'         : $('input[name=email]').val(),
            'subject'       : $('input[name=subject]').val(),
            'message'       : $('textarea[name=message]').val(),
            'auth-token'    : $('input[name=auth-token]').val()
        };

        // process the form
        $.ajax({
            type        : 'POST',
            url         : '/send', // the url where we want to POST
            data        : formData,
            dataType    : 'json', // what type of data do we expect back from the server
            encode      : true
        })
            .done(function(data) {
                // here we handle a successful submission
                console.log(data);
            });

        // stop the form from refreshing the page
        e.preventDefault();
    });

});
```

### CURL example
```shell
curl -i -X POST -H "Content-Type: application/json" -H "X-Dispatch-Subject: cmd email" -d '{ "auth-token":"qasZ1z6HfVPRCq1D0GQUpVB8", "name":"anon", "email":"test@dispatch.com", "message":"Hello!"}' http://dispatch:7070/send
```

## Documentation

This documentation can be found at github.com/gesquive/dispatch

## License

This package is made available under an MIT-style license. See LICENSE.

## Contributing

PRs are always welcome!
