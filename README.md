# Clipper-API

API microservice of Clipper CI\CD.

## Overview

This microservice is used to provide RESTfull API of CI\CD system, as well as to handle webhooks from Github and create CI jobs.

## Installation
1. Build api binary using `go build` or use docker image.
2. Don't forget to securily provide credentials for initial admin account, JWT secret, Postgres etc.

## Command line arguments
Call api binary with `help` argument to see built-in help.
Call api binary with `start` argument and next parameters.

| Parameter (short)     | Default                           | Usage                                                     |
|-----------------------|-----------------------------------|-----------------------------------------------------------|
| --port (-p)           | 8080                              | Application port                                          |
| --rabbitmq (-r)       | amqp://guest:guest@localhost:5672 | rabbitmq connection URL                                   |
| --postgresAddr (-a)   | postgres:5432                     | PostgreSQL address                                        |
| --db (-d)             | clipper                           | PostgreSQL database to use                                |
| --user (-u)           | clipper                           | PostgreSQL database user                                  |
| --pass (-c)           | clipper                           | PostgreSQL user's password                                |
| --adminlogin (-l)     | admin                             | First admin account login                                 |
| --adminpass (-x)      | admin                             | First admin account password                              |
| --adminlogin (-l)     | admin                             | First admin account login                                 |
| --jwt (-j)            | veryverysecret                    | JWT token's secret                                        |
| --verbose (-v)        | false                             | Show debug level logs                                     |