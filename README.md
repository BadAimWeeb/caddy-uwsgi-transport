# caddy-uwsgi-transport

This module adds [uwsgi](https://uwsgi-docs.readthedocs.io/en/latest/Protocol.html) reverse proxying support to Caddy.

ID: `http.reverse_proxy.transport.uwsgi`

## Installation

By using [`xcaddy`](https://caddyserver.com/docs/build#xcaddy)

```sh
xcaddy build \
    --with github.com/BadAimWeeb/caddy-uwsgi-transport
```

## Usage

### Caddyfile

```caddyfile
reverse_proxy [<matcher>] [<upstreams...>] {
	transport uwsgi {
    uwsgi_param <key> <value> # in case you need to set uwsgi params (for example UWSGI_SCRIPT)
  }
}
```

### JSON

```json
{
  "apps": {
    "http": {
      "servers": {
        "": {
          "routes": [
            {
              "handle": [
                {
                  "handler": "reverse_proxy",
                  "transport": {
                    "protocol": "uwsgi",
                    "uwsgi_params": {
                      "<key>": "<value>"
                    }
                  },
                  "upstreams": [{ "dial": "<upstream>" }]
                }
              ]
            }
          ]
        }
      }
    }
  }
}
```

## Copyright

This module is licensed under the [Apache 2.0 License](LICENSE).

Copyright ©️ 2024 BadAimWeeb.<br />Original author copyright ©️ 2023 Xinhe Wang.