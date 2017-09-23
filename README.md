# CASPER
_Configuration Automation for Safe and Painless Environment Releases_

[![Build Status](https://secure.travis-ci.org/miracl/casper.png?branch=master)](https://travis-ci.org/miracl/casper?branch=master)
[![Coverage Status](https://coveralls.io/repos/miracl/casper/badge.svg?branch=master&service=github)](https://coveralls.io/github/miracl/casper?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/miracl/casper)](https://goreportcard.com/report/github.com/miracl/casper)

## Description

Casper is a simple tool for managing configurations as code where the structure of the configurations is separated from the values themselves. That way you can use the same structure for different environments. The structure of the configurations is described in a template file. Overall the tool combines the template file with all the values, serializes it to a format applicable for the configuration storage and pushes the changes to the storage.

## Installation

### From source 

```
go get -u github.com/miracl/casper/cmd/casper
```

### From [GitHub releases](https://github.com/miracl/casper/releases)

## Example
```
cd examples/
casper build
```
* [template.yaml](/example/template.yaml)
* [config.yaml](/example/config.yaml)
* [source.yaml](/example/source.yaml)
* [expected output](/example/output.yaml)

## Usage

All configurations can be given on the command line, with file or with environment variables. Check `casper -h` for full list.

* **template** - The template file is a golang template. The end product of the template file and the values should be of a format applicable for the configuration storage (e.g: json, yaml for key/value stores)
* **sources** - Sources are the thing containing the keys for the template. Sources is a list. Currently there are 2 available:
	* Config source is a list of key/value pairs directly in the configuration file. Check ([config.yaml](/example/config.yaml)) for examples. 
		```
		sources:
		- key1=val1
		- key2=val2
		```
	* File source defines key/value pairs in a file. Currently supported formats are `json` and `yaml`.
		```
		sources:
		- file://source.yaml
		```
* **storage** - Storage is the system that Casper menages. Currently there are 2 available:
	* Consul.
		```
		storage: consul
		consul-addr: http://172.17.0.1:8500/?token=acl_token&ignore=_ignore
		```
		* ignore - keys given the value of this setting in configuration will be ignored by Casper. The default such value is `_ignore`
	* File
		```
		storage: file
		file-path: output.yaml
		```
