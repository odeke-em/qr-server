# qr-server
QR as a service server


### Env variables

```shell
export MONGOHQ_URI=<MongoHqURI> # Required
export DBNAME="aDBname" # Required
export RESTRICT_DOMAINS=true # or false -- Optional
```

### Domain restriction
When RESTRICT\_DOMAINS is set, generic paths "/qr" won't be exposed.
