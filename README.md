# seneye-exporter

seneye-exporter provides a webserver which listens for pushes from the [Seneye Connect App (SCA)](https://sca.seneye.com/) or [Seneye Web Server (SWS)](https://www.seneye.com/store/seneye-web-server.html) and publishes the results as prometheus metrics.

## Usage
```
Usage:
  seneye-exporter [flags]

Flags:
      --config string        config file
  -h, --help                 help for seneye-exporter
      --lde-port uint16      Port for LDE server (default 8080)
      --lde-secret strings   Secret used to validate LDE message authenticity. --lde-secret may be specified
                             multiple times if paired with the SUD ID. (ex. --lde-secret=DEFAULT_SECRET, or
                             --lde-secret=EXAMPLE_SUD_ID=SECRET1 --lde-secret=OTHER_SUD_ID=SECRET2)
      --log-format string    log format: "json", "text" (default "text")
      --log-level string     log level: "trace" "debug" "info" 
                             "warn" "error" "fatal" "panic" (default "debug")
      --prom-port uint16     Port for prometheus metrics server (default 9090)
```

## Kubernetes
```
kubectl create namespace seneye-exporter
kubectl create secret -n seneye-exporter generic --from-literal=LDE_SECRET=XXXXXXXX seneye-exporter
```