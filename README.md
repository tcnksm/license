# license

[![GitHub release](http://img.shields.io/github/release/tcnksm/license.svg?style=flat-square)][release]
[![MIT License](http://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)][license]
[![Go Documentation](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)][godocs]

[release]: https://github.com/tcnksm/license/releases
[license]: https://github.com/tcnksm/license/blob/master/LICENSE
[godocs]: http://godoc.org/github.com/tcnksm/license

`license` is a simple command line tool to generate LICENSE file you prefer. It fetches one from [Github API](https://developer.github.com/v3/licenses/) and ask you place to replace. It's good point for you to start your OSS product.

## Usage

To generage `LICENSE` file, you just provide `KEY` name of LICENSE you want,

```bash
$ license [option] [KEY]
```

To check avairable `LICESE` file, you can see all of them by `-list` options

```bash
$ license -list
```

If you don't provide specific `KEY`, `license` will ask you which `LICENSE` is good for you.

If you feel difficulty to choose `LICENSE` for your project, `license` command provide you a way to choose `LICENSE`
you need by `-choose` option, it will ask you to choose `LICESE` like [choosealicense.com/](http://choosealicense.com/) ,

```bash
$ license -choose
```

To see more usage, use `-help` option

## Install 

Binaries for your platform are provided, install it from [Relase page]().

If you use OSX and [homebrew]() for your package manager, you can use fomula for this project,

```bash
$ tap tcnksm/license
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

[tcnksm](https://github.com/tcnksm)
