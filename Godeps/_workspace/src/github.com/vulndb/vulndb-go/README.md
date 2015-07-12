# vulndb-go

Go SDK to access the [vulnerability database](https://github.com/vulndb/data)

## Installation

```
go get github.com/vulndb/vulndb-go
```

## Usage

Load vulndb data from binary data and link to package

```
package main

import (
	"github.com/vulndb/vulndb-go/bindata"
	"log"
	"fmt"
)

func main() {
	vulns, err := bindata.LoadFromBin()
	if err != nil {
		log.Fatal(err)
	}
	for _, vuln := range vulns.FilterBySeverity("high") {
		fmt.Printf("%2d. %-52.50s [%s]\n", vuln.Id, vuln.Title, vuln.Severity)
	}
}
```

Contributing
============
Send your [pull requests](https://help.github.com/articles/using-pull-requests/)
with improvements and bug fixes, making sure that all tests ``PASS``:

```

    $ cd vulndb-go
    $ make tools # update required tools
    $ make
    Test packages
    PASS
    ok  	github.com/vulndb/vulndb-go/bindata	0.026s
    Run vet
    Check formats
```

Updating the database
=====================
This package embeds the [vulnerability database](https://github.com/vulndb/data)
in the ``bindata`` directory. To update the database with new information
follow these steps:

```
    # Update the database
    make update-db

```