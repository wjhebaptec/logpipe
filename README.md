# logpipe

A structured log aggregator with filtering and forwarding rules defined in a simple YAML config.

---

## Installation

```bash
go install github.com/yourusername/logpipe@latest
```

Or build from source:

```bash
git clone https://github.com/yourusername/logpipe.git && cd logpipe && go build -o logpipe .
```

---

## Usage

Define your rules in a `logpipe.yaml` config file:

```yaml
inputs:
  - type: stdin
  - type: file
    path: /var/log/app.log

filters:
  - field: level
    match: error

outputs:
  - type: stdout
  - type: file
    path: /var/log/errors.log
  - type: http
    url: https://logs.example.com/ingest
```

Then run logpipe:

```bash
logpipe --config logpipe.yaml
```

You can also pipe logs directly:

```bash
cat app.log | logpipe --config logpipe.yaml
```

---

## Configuration

| Key | Description |
|---|---|
| `inputs` | Sources to read structured logs from |
| `filters` | Rules to match or exclude log entries |
| `outputs` | Destinations to forward matching logs |

---

## License

MIT © [yourusername](https://github.com/yourusername)