# GCServe

Serve files from Google Cloud Storage.

Features:

+ Basic HTTP auth.

## Usage

```bash
go build -o gcserve.bin .
./gcserve.bin --bucket=test-bucket --username=test-user --password=test-pass --debug
```

## Alternatives

+ [gcsproxy](https://github.com/daichirata/gcsproxy)
+ [weasel](https://github.com/google/weasel)
