# GCServe

Serve files from a [Google Cloud Storage](https://cloud.google.com/storage) bucket.

There are already a few alternatives (namely, [gcsproxy](https://github.com/daichirata/gcsproxy) and [weasel](https://github.com/google/weasel)). Also, GCP allows to [make all bucket files public](https://cloud.google.com/storage/docs/hosting-static-website). The main difference is that GCServe provides [basic HTTP auth](https://en.wikipedia.org/wiki/Basic_access_authentication).

## Using as PyPI server

A possible use-case for GCServe is to host a private [PyPI](https://pypi.org/) instance. All you need is:

1. Download packages from pypi.org using [pip download](https://pip.pypa.io/en/stable/reference/pip_download/).
1. Upload packages into the bucket using [gsutil rsync](https://cloud.google.com/storage/docs/gsutil/commands/rsync).
1. Generate static index using [dumb-pypi](https://github.com/chriskuehl/dumb-pypi).
1. Serve the bucket with GCServe.

This pipeline is much faster, smaller, and more reliable than a more dynamic solution, like [pypicloud](https://github.com/stevearc/pypicloud).

## Usage

Build and run using [Go](https://golang.org/) compiler:

```bash
go build -o gcserve.bin .
./gcserve.bin \
    --bucket=test-bucket \
    --username=test-user \
    --password=test-pass \
    --debug
```

Build and run using [Docker](https://www.docker.com/):

```sh
sudo docker build -t gcserve:latest .
sudo docker run \
    -v /path/to/google/credentials.json:/mnt/cred.json \
    -it gcserve:latest \
    --bucket=test-bucket \
    --username=test-user \
    --password=test-pass \
    --debug
```
