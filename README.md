# CASPER
_Configuration Automation for Safe and Painless Environment Releases_

[![Build Status](https://secure.travis-ci.org/miracl/casper.png?branch=master)](https://travis-ci.org/miracl/casper?branch=master)
[![Coverage Status](https://coveralls.io/repos/miracl/casper/badge.svg?branch=master&service=github)](https://coveralls.io/github/miracl/casper?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/miracl/casper)](https://goreportcard.com/report/github.com/miracl/casper)

## Description

CASPER is a simple tool for managing software configuration.

Configuration structure is separated from the values it holds. This way you can use the same structure for different environments or encode the same values in different formats.

The two main units CASPER uses are template files, describing structure, and configuration files, containing the values to use.

In a nutshell, the tool combines the template files with the values into a serialized format compatible with the target storage system. It then pushes the changes to the storage.

## Installation

```
go get -u github.com/miracl/casper
```

## Example
```
cd examples/
casper build
```
* [template.yaml](/example/template.yaml)
* [config.yaml](/example/config.yaml)
* [expected output](/example/output.yaml)

## config.yaml

The `config.yaml` file sets up the way Casper should behave. Study the example file listing below and its breakdown afterwards to understand how to configure Casper.

```
template: <path to a template file>
format: <format of the template file>
sources:
  - type: <config|file>
    vals:
      <key1>: <value1>
      <key2>: <value2>
      <keyn>: <valuen>
storage:
  type: <consul|file>
  config:
    <config key>: <config value>
```

**template** - the template file is a [golang template](https://golang.org/pkg/text/template/). The end product of the template file and the values will be of a format applicable for the configuration storage (e.g: json, yaml for key-value stores)

**format** - the format of the template, for example "yaml". It depends on storage configuration.

**sources** - the key-value pairs applied to the template. _Sources_ is a list of sources defined with `type`, listing keys that are specific to the particular source. Currently there are two types of sources available:

* `config` - key-value pairs are placed directly in the configuration file
	* `vals` - a list of key-value pairs
* `file` - the key-value pairs are stored in a separate file, allowing you to separate Casper's configuration from the actual data
	* `format` - the format of the file containing the values (valid formats are "json" and "yaml")
		* `file` - the path to file containing key-value pairs

**storage** - Storage is the system that Casper manages. Storage will have two keys - `type` and `config`, where `type` specifies the type of the storage being used, while `config` contains configuration for the storage itself. There are two types of storage supported currently:
* consul - [Consul](https://www.consul.io/) (accepted formats: json, yaml)
	* addr - address of a Consul instance e.g: `localhost:8500`
* file - File (accepted format: string)
	* path - path to the file


