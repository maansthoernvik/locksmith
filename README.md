# Locksmith <!-- omit in toc -->

![build](https://github.com/maansthoernvik/locksmith/actions/workflows/build.yml/badge.svg)
[![codecov](https://codecov.io/gh/maansthoernvik/locksmith/graph/badge.svg?token=6MrGbVWC5b)](https://codecov.io/gh/maansthoernvik/locksmith)

- [1. Install](#1-install)
- [2. How to run](#2-how-to-run)
  - [2.1. The locksmith server](#21-the-locksmith-server)
    - [2.1.1. Locksmith server environment variables](#211-locksmith-server-environment-variables)
    - [2.1.2. Advanced configuration options](#212-advanced-configuration-options)
  - [2.2. The command line utility](#22-the-command-line-utility)
- [3. How to use the locksmith code as a library](#3-how-to-use-the-locksmith-code-as-a-library)


Locksmith provides a simple way to obtain shared locks between applications.

This project provides both server software, a command line utility, and a sample client. The protocol package can also be used to write custom client software.

## 1. Install

The locksmith server can be installed in two different ways. To get the server container image, run:

```bash
docker pull ghcr.io/maansthoernvik/locksmith:latest
```
*You can browse available versions here: https://github.com/maansthoernvik/locksmith/pkgs/container/locksmith*

Run `go install` to instead get the server binary, make sure you have set either `GOPATH` or `GOBIN`.

```bash
go install github.com/maansthoernvik/locksmith/cmd/locksmith@v0.1.0
```

The command line utility must be installed via `go install`.

```bash
go install github.com/maansthoernvik/locksmith/cmd/locksmithctl@v0.1.0
```

## 2. How to run

### 2.1. The locksmith server

Both the binary and the docker container have configuration options that are consumed via environment variables. All variables are namespaced and prefixed with `LOCKSMITH_`.

#### 2.1.1. Locksmith server environment variables

- `LOCKSMITH_LOG_LEVEL`: If set, the given value MUST be either `DEBUG`, `INFO`, `WARNING`, `ERROR`, or `CRITICAL` (default: `WARNING`)
- `LOCKSMITH_PORT`: The port where the locksmith server is reachable (default: `9000`)
- `LOCKSMITH_TLS`: If set to `true`, TLS is enabled for the locksmith server (default: `false`). When enabled, both `LOCKSMITH_TLS_CERT_PATH` and `LOCKSMITH_TLS_KEY_PATH` must be provided or locksmith will panic and stop on start
- `LOCKSMITH_TLS_CERT_PATH`: Absolute path to the server´s certificate
- `LOCKSMITH_TLS_KEY_PATH`: Absolute path to the server´s private key
- `LOCKSMITH_TLS_REQUIRE_CLIENT_CERT`: When set to `true` (default: `false`), client connections will have their certificates validated against the client CA certificate. You must provide `LOCKSMITH_TLS_CLIENT_CA_CERT_PATH` when this variable is set
- `LOCKSMITH_TLS_CLIENT_CA_CERT_PATH`: Absolute path to the client CA certificate

#### 2.1.2. Advanced configuration options

These options are available to enable optimizations for systems where locksmith is expected to handle very high loads. These parameters are not recommended to be changed at all for most use cases as the defaults are meant to be good enough for 99% of cases. Ensure that any changes to these values are preceded by rigorous load testing in the intended environment.

- `LOCKSMITH_Q_TYPE`: Testing utility, there are only two options: `single` and `multi` (default: `multi`). The `single` option completely removes concurrency, making locksmith inefficient at handling higher throughput, but easier to test
- `LOCKSMITH_Q_CONCURRENCY`: Only applicable for `multi` type queueing, sets the number of go-routines serving incoming requests (default: `10`)
- `LOCKSMITH_Q_CAPACITY`: Only applicable for `multi` type queueing, sets the size of each serving go-routines work queue (default: `100`)

### 2.2. The command line utility

## 3. How to use the locksmith code as a library


