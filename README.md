> This package is a work in progress

# Ringslice

[![go workflow](https://github.com/chronohq/ringslice/actions/workflows/go.yml/badge.svg)](https://github.com/chronohq/ringslice/actions/workflows/go.yml)
[![mit license](https://img.shields.io/badge/license-MIT-green)](/LICENSE)

A generic ring buffer backed by a Go slice, with iter.Seq iteration and callback hooks for production use cases.

For a traditional circular doubly-linked list, see Go's standard [container/ring](https://pkg.go.dev/container/ring) package.
