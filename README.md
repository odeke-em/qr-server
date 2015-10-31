# qr-demo-server
QR share demo server


### Env variables

```shell
export MONGOHQ_URI=<MongoHqURI> # Needed
export RESTRICT_DOMAINS=true # or false -- Optional
```

### Domain restriction
When RESTRICT\_DOMAINS is set, generic paths "/qr" won't be exposed.
