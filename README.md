# Azure exporter

[![Build Status](https://travis-ci.org/iamseth/azure_exporter.svg)](https://travis-ci.org/iamseth/azure_exporter)
[![GoDoc](https://godoc.org/github.com/iamseth/azure_exporter?status.svg)](http://godoc.org/github.com/iamseth/azure_exporter)
[![Report card](https://goreportcard.com/badge/github.com/iamseth/azure_exporter)](https://goreportcard.com/badge/github.com/iamseth/azure_exporter)

Prometheus exporter for Azure metrics using the Azure Resource Manager API. Currently, it only supports VPN connections. Eventually, it will support storage account metrics as well.

Microsoft limits API reads to 15000 per hour. Keep this in mind when setting the scrape interval. See [Azure subscription and service limits, quotas, and constraints](https://azure.microsoft.com/en-us/documentation/articles/azure-subscription-service-limits/) for more details.

## Install

```bash
go get -u github.com/iamseth/azure_exporter
```

## Usage
```bash
Usage of azure_exporter:
  -credentials-file string
    	Specify the JSON file with the Azure credentials. (default "~/.azure/credentials.json")
  -listen-address string
    	The address to listen on for HTTP requests. (default ":9080")
  -log.format value
    	If set use a syslog logger or JSON logging. Example: logger:syslog?appname=bob&local=7 or logger:stdout?json=true. Defaults to stderr.
  -log.level value
    	Only log messages with the given severity or above. Valid levels: [debug, info, warn, error, fatal]. (default info)
  -web.telemetry-path string
    	Path under which to expose metrics. (default "/metrics")
```

## Binary releases

Pre-compiled versions may be found in the [release section](https://github.com/iamseth/azure_exporter/releases).
