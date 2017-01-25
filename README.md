# CASPER
_Configuration Automation for Safe and Painless Environment Releases_

[![Build Status](https://secure.travis-ci.org/miracl/casper.png?branch=master)](https://travis-ci.org/miracl/casper?branch=master)
[![Coverage Status](https://coveralls.io/repos/miracl/casper/badge.svg?branch=master&service=github)](https://coveralls.io/github/miracl/casper?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/miracl/casper)](https://goreportcard.com/report/github.com/miracl/casper)

## Description

Casper is a simple tool for managing configurations as code where the structure of the configurations is separated from the values themselves. That way you can use the same structure for different environments. The structure of the configurations is described in a template file. Overall the tool combines the template file with all the values, serializes it to a format applicable for the configuration storage and pushes the changes to the storage.

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

* **template** - The template file is a golang template. The end product of the template file and the values should be of a format applicable for the configuration storage (e.g: json, yaml for key/value stores)
* **format** - Format of the template. It depends on the storage configures.
* **sources** - Sources are the thing containing the keys for the template. Sources is a list of sources defined with `type` and keys that are specific to each particular source. Currently there are 2 available:
	* config - key/value pairs directly in the configuration file
		* vals - list of key/value pairs
	* file - key/value pairs in file
		* format - format of the file (json, yaml)
		* file - path to file containing key/value pairs
* **storage** - Storage is the system that Casper menages. Storage will have two keys - `type` and `config` where `config` contains all the configurations for the storage specified by `type`. Currently there are 2 available:
	* consul - Consul (Formats: json, yaml)
		* addr - address of the consul instance e.g: `localhost:8500`
	* file - File (Formats: string)
		* path - path to the file
