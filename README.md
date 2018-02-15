# license

[![GitHub release](http://img.shields.io/github/release/tcnksm/license.svg?style=flat-square)][release]
[![MIT License](http://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)][license]
[![Go Documentation](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)][godocs]

[release]: https://github.com/tcnksm/license/releases
[license]: https://github.com/tcnksm/license/blob/master/LICENSE
[godocs]: http://godoc.org/github.com/tcnksm/license

`license` is a simple command line tool to generate LICENSE file you want to use. It fetches one from [Github API](https://developer.github.com/v3/licenses/). You can also choose LICENSE like [choosealicense.com](http://choosealicense.com/). 

## Demo

You can select a LICENSE from available list,

![](http://g.recordit.co/IlnUBhCUHX.gif)

You can just provide key name,

![](http://g.recordit.co/FRKXgTvrml.gif)

If you feel difficulty to choose LICENSE, you can do it like [choosealicense.com](http://choosealicense.com/),

![](http://g.recordit.co/2MZs3RTnSd.gif)

## Usage

To generate LICENSE file, you just provide `KEY` name of LICENSE you want,

```bash
$ license [option] [KEY]
```

To check available `LICENSE` file and its `KEY`, you can see all of them by `-list` option, 

```bash
$ license -list
```

If you don't provide specific `KEY`, `license` will ask you to select one from list.

To choose LICENSE like [choosealicense.com](http://choosealicense.com/),

```bash
$ license -choose
```

To see more usage, use `-help` option

## Install 

Binaries for your platform are provided, install it from [Release page](https://github.com/tcnksm/license/releases).

If you use OSX and [homebrew](http://brew.sh/) for your package manager, you can use formula for this project,

```bash
$ brew tap tcnksm/license
$ brew install license
```

## Contribution

1. Fork ([https://github.com/tcnksm/license/fork](https://github.com/tcnksm/license/fork))
1. Create a feature branch
1. Commit your changes
1. Rebase your local changes against the master branch
1. Run test suite with the `go test ./...` command and confirm that it passes
1. Run `gofmt -s`
1. Create a new Pull Request

To install latest version of `license`, use `go get`:

```bash
$ go get -d github.com/tcnksm/license
```


## Author

[Taichi Nakashima](https://github.com/tcnksm)
