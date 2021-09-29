# chronon

`chronon` is a testable time package.

[![Build Status](https://github.com/xmidt-org/chronon/workflows/CI/badge.svg)](https://github.com/xmidt-org/chronon/actions)
[![codecov.io](http://codecov.io/github/xmidt-org/chronon/coverage.svg?branch=main)](http://codecov.io/github/xmidt-org/chronon?branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/xmidt-org/chronon)](https://goreportcard.com/report/github.com/xmidt-org/chronon)
[![Apache V2 License](http://img.shields.io/badge/license-Apache%20V2-blue.svg)](https://github.com/xmidt-org/chronon/blob/main/LICENSE)
<!--
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=xmidt-org_chronon&metric=alert_status)](https://sonarcloud.io/dashboard?id=xmidt-org_chronon)
-->
[![GitHub release](https://img.shields.io/github/release/xmidt-org/chronon.svg)](CHANGELOG.md)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/xmidt-org/chronon)](https://pkg.go.dev/github.com/xmidt-org/chronon)

## Table of Contents

- [Overview](#overview)
- [Code of Conduct](#code-of-conduct)
- [Install](#install)
- [Contributing](#contributing)

## Overview

`chronon` aims to make concurrent, time-related `golang` code easier to test.  In particular, `chronon` avoids having package-level state or a "test mode" that unit tests use to drive timers, tickers, etc.

## Code of Conduct

This project and everyone participating in it are governed by the [XMiDT Code Of Conduct](https://xmidt.io/docs/community/code_of_conduct/). 
By participating, you agree to this Code.

## Install

```
go get -u github.com/xmidt-org/chronon
```

## Contributing

Refer to [CONTRIBUTING.md](CONTRIBUTING.md).
