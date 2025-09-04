# sqwat

[![Release](https://img.shields.io/github/v/release/donderom/sqwat.svg?style=flat-square&color=6e54da)](https://github.com/donderom/sqwat/releases)
[![Build](https://img.shields.io/github/actions/workflow/status/donderom/sqwat/build.yml?style=flat-square&logo=github)](https://github.com/donderom/sqwat/actions/workflows/build.yml)
[![ReportCard](https://goreportcard.com/badge/github.com/donderom/sqwat?style=flat-square)](https://goreportcard.com/report/donderom/sqwat)
[![License](https://img.shields.io/badge/license-MIT-463494?style=flat-square)](https://github.com/donderom/sqwat/blob/main/LICENSE)

<p align="center">
<img src="logo.svg" width="128" align="center" alt="The sqwat logo">
</p>

A TUI editor for files in the [Stanford Question Answering Dataset](https://rajpurkar.github.io/SQuAD-explorer/) (SQuAD) format.

* Preview SQuAD files
* Modify any part of the dataset (delete, edit, create, etc.)
* Validation for common issues
* Supports both SQuAD versions 1.1 and 2.0
* Full-text search across all fields
* Highlights answers within the context with validation
* Accumulated warnings with navigation

## Installation

Install `sqwat` with Go:

```sh
GOEXPERIMENT=jsonv2 go install github.com/donderom/sqwat@latest
```

Or download the latest [release](https://github.com/donderom/sqwat/releases) for your system and architecture.

## Usage

Run `sqwat` with a file in the SQuAD format as its argument:

```sh
sqwat train-v2.0.json
```

The original SQuAD dataset files can be found [here](https://github.com/rajpurkar/SQuAD-explorer/tree/master/dataset).

---

*Built with [bubblon](https://github.com/donderom/bubblon).*
