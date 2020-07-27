![Neuron Logo](logo.svg)

# Neuron Postgres [![Go Report Card](https://goreportcard.com/badge/github.com/neuronlabs/neuron-postgres)](https://goreportcard.com/report/github.com/neuronlabs/neuron-postgres) [![GoDoc](https://godoc.org/github.com/neuronlabs/neuron?status.svg)](https://godoc.org/github.com/neuronlabs/neuron-postgres) [![Build Status](https://travis-ci.com/neuronlabs/neuron-postgres.svg?branch=master)](https://travis-ci.com/neuronlabs/neuron-postgres) ![License](https://img.shields.io/github/license/neuronlabs/neuron-postgres.svg) 

`neuron-postgres` is the `github.com/neuronlabs/neuron` postgres repository driver.


- [What is Neuron-PQ](#what-is-neuron-postgres)
- [Installation](#installation)
- [QuickStart](#quick-start)
- [Docs](#docs)
- [Examples](#examples)
- [Options](#options)
- [neuron](https://github.com/neuronlabs/neuron)

## What is Neuron-Postgres

 Neuron-PQ is the PostgreSQL ORM Repository driver for the `github.com/neuronlabs/neuron`. 

## Installation

`go get github.com/neuronlabs/neuron-postgres`

Neuron-postgres is the extension for the [neuron](https://docs.neuronlabs.io/neuron) requires `github.com/neuronlabs/neuron` root package.

## QuickStart


1. Prepare or read the Config Repository - `github.com/neuronlabs/neuron/config` by creating structure or by using viper  `github.com/spf13/viper` and create the neuron controller
The `DriverName` variable (`driver_name` in config) should be set to `postgres` in order to match the `neuron-postgres` Repository and Factory

`models.go`
```go
package main

type User struct {
    ID      int   
    Name    string
    Surname string
    Cars    []*Car `neuron:"foreign=OwnerID"`
}

type Car struct {
    ID              int    
    RegisterNumber  string  `db:"unique"`
    Owner           *User   `neuron:"type=relation;foreign=OwnerID"`
    OwnerID         int     `db:"notnull"`
}
```


`main.go`
```go
package main

import (
    "context"

    "github.com/neuronlabs/neuron"
    "github.com/neuronlabs/neuron/config"
    "github.com/neuronlabs/neuron/log"
    // import postgres repository driver
    _ "github.com/neuronlabs/neuron-extensions/repository/postgres"
)


func main(){
    cfg := &config.Repository{
        DriverName: "postgres",
        Host: "localhost",
        Port: 5432,
        DBName:"dbname",
        Password: "my-password",
        Username: "username",
    }
    // Register the neuron-postgres repository for you application by initializing the repository package
    // Add more custom repository configurations for different models if you 
    // need different connection or credentials
    err := neuron.RegisterRepository("mypostgres", cfg)
    if err != nil { 
        log.Fatal(err)
    }    
    
    if err = neuron.DialAll(context.Background()); err != nil {
        log.Fatal(err)   
    }
    
    // Register the models 
    if err = neuron.RegisterModels(&User{}, &Car{}); err != nil {
        log.Fatal(err)
    }
    
    // We can create and update current tables for given model by auto migrating it.
    if err = neuron.MigrateModels(context.Background(), &User{}); err != nil {
        log.Fatal(err)    
    }

    // Use the models within the Gateway API or just as an ORM models.
    user := &User{
        Name: "Mathew",
        Surname: "Smith",
    }

    err = neuron.Query(user).Create()
    if err != nil {
        log.Fatal(err)
    }
    log.Infof("Inserted user: %v successfully", user)
}
```

## Examples

`config.yml`
```json
{
    "repositories": {
        "postgres": {
            "driver_name": "postgres",
            "host": "localhost",
            "port": "5432",
            "dbname": "testing",
            "username": "testing",
            "password": "testing"
        }
    },
    "default_repository_name": "postgres"
}
```


## Docs

- neuron: https://docs.neuronlabs.io/neuron
- Neuron-PQ: https://docs.neuronlabs.io/neuron-postgres
- GoDoc: https://godoc.org/github.com/neuronlabs/neuron-postgres
- Project Neuron: https://docs.neuronlabs.io
