A temporary workaround to get allocated resources from each Nomad client into graphite.

This is a workaround for https://github.com/hashicorp/nomad/issues/2072

```
Usage of ./nomad_alloc_metrics:
  -addr string
        The address of the Nomad server (default "http://127.0.0.1:4646")
  -graphite-addr string
        host and port of carbon server (default "127.0.0.1:2003")
  -tls-ca-cert string
        TLS CA cert to use to validate the Nomad server certificate
  -tls-cert string
        TLS certificate to use when connecting to Nomad
  -tls-insecure
        Whether or not to validate the server certificate
  -tls-key string
        TLS key to use when connecting to Nomad
```
