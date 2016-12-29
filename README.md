# Build

```bash
go get -d -v -t

go test --cover

# Poor coverage. Skipped tests for the main function.

go build -v -o foo-protocol
```

# Run

```bash
# Change darwin for linux if needed

./orig/server-darwin -listen=:8001

./foo-protocol -help

./foo-protocol

# Change darwin for linux if needed

./orig/client-darwin -connect=localhost:8002

kill -SIGUSR1 $(ps aux | grep "foo\-protocol" | awk '{print $2}')

# Did not have time to implement the rest of metrics. But...

curl "localhost:8003/metrics"

# Plug it into Prometheus, and you'll have the metrics
```
