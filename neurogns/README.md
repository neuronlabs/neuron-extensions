![Neuron Logo](logo_teal.svg)

Neuro**G**o**N**esi**S** (`neurogns`) is the **golang** based generator for the `neuron` framework.

# Neuro**Go**Nesis [![Build Status](https://travis-ci.com/neuronlabs/neurogonesis.svg?branch=master)](https://travis-ci.com/neuronlabs/neurogonesis) ![GitHub](https://img.shields.io/github/license/neuronlabs/neurogonesis)

 is a part of the Neuron framework - [github.com/neuronlabs/neuron](https://github.com/neuronlabs/neuron).

It is a CLI generator that enhances your models with implementing required methods and structures.

## Usage 

This generator should be used to generate your models methods and database collection access 
by providing the `go:generate` comments within your model files.

### Models

`neurogonesis` can generate all required methods for the `neuron/mapping` model interfaces for all your models.

In order to scan and generate paste following command in your models: 
`//go:generate neurogonesis models --format=goimports --single-file .`

Then execute command `go generate` within the directory your files are located.

The `--single-file` flag stores all models methods in the single file.

If you want to scan and generate files only for specified types, use the flag:
`--type=Type1,Type2` - which would generate model methods only for `Type1` and `Type2` 

If you need to exclude some types that might be recognized as a model provide a flag:
`--exclude=Type1,Type2` which would not tread `Type1` and `Type2` as models.

The `--format=goimports` formats file using `goimports` formatter.

For more information type`neurogonesis models --help` .

### Collections

Neuron generator can generate `Collections` for each scanned model. Collections are structures that contains non-interface
database model queries as well as customized functions and query builders. 

Collections are designed to execute queries easily, without the necessity of interface conversions.

In order to scan and generate paste following command in your models: 
`//go:generate neurogonesis collections --format=goimports --single-file .`

Then execute command `go generate` within the directory your files are located.

The `--single-file` flag stores all models methods in the single file.

If you want to scan and generate files only for specified types, use the flag:
`--type=Type1,Type2` - which would generate model collections only for `Type1` and `Type2`

If you need to exclude some types that might be recognized as a model provide a flag:
`--exclude=Type1,Type2` which would not tread `Type1` and `Type2` as models.

The `--format=goimports` formats file using `goimports` formatter.

For more information type`neurogonesis collections --help` .

