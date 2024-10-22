# Registrator

This is a fork of <https://github.com/gliderlabs/registrator>. \
A Service registry bridge for Docker.

Registrator automatically registers and deregisters services for any Docker
container by inspecting containers as they come online. Registrator
supports pluggable service registries, which currently includes
[Consul](http://www.consul.io/), [etcd](https://github.com/coreos/etcd) and
[SkyDNS 2](https://github.com/skynetservices/skydns/).

## Using Registrator

The quickest way to see Registrator in action is our
[Quickstart](docs/user/quickstart.md) tutorial. Otherwise, jump to the [Run
Reference](docs/user/run.m) in the User Guide. Typically, running Registrator
looks like this:

    $ docker run -d \
        --name=registrator \
        --net=host \
        --volume=/var/run/docker.sock:/tmp/docker.sock \
        gliderlabs/registrator:latest \
          consul://localhost:8500

## CLI Options
```
Usage of /bin/registrator:
  /bin/registrator [options] <registry URI>

  -cleanup=false: Remove dangling services
  -deregister="always": Deregister exited services "always" or "on-success"
  -explicit=false: Only register containers which have SERVICE_NAME label set
  -internal=false: Use internal ports instead of published ones
  -ip="": IP for ports mapped to the host
  -resync=0: Frequency with which services are resynchronized
  -retry-attempts=0: Max retry attempts to establish a connection with the backend. Use -1 for infinite retries
  -retry-interval=2000: Interval (in millisecond) between retry-attempts.
  -tags="": Append tags for all registered services
  -ttl=0: TTL for services (default is no expiry)
  -ttl-refresh=0: Frequency with which service TTLs are refreshed
```

## License

MIT
