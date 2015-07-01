# license

[![GitHub release](http://img.shields.io/github/release/tcnksm/license.svg?style=flat-square)][release]
[![MIT License](http://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)][license]
[![Go Documentation](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)][godocs]

[release]: https://github.com/tcnksm/license/releases
[license]: https://github.com/tcnksm/license/blob/master/LICENSE
[godocs]: http://godoc.org/github.com/tcnksm/license

`license` is a simple command line tool to generate LICENSE file you want. It fetches it from [Github API](https://developer.github.com/v3/licenses/). 

## Usage

```bash
$ license [option] [KEY]
```

## Install

To install, use `go get`:

```bash
$ go get -d github.com/tcnksm/license
```

## Contribution

1. Fork ([https://github.com/tcnksm/license/fork](https://github.com/tcnksm/license/fork))
1. Create a feature branch
1. Commit your changes
1. Rebase your local changes against the master branch
1. Run test suite with the `go test ./...` command and confirm that it passes
1. Run `gofmt -s`
1. Create a new Pull Request

## Author

[tcnksm](https://github.com/tcnksm)
