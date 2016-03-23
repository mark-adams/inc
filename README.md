# Overview

inc is a simple microservice for keeping track of incrementing counters

# Installation

```
go install github.com/mark-adams/inc
```

# Usage
## Create a new counter
```
$ http POST http://localhost:3000/new  

HTTP/1.1 201 Created
Content-Length: 32
Content-Type: text/plain; charset=utf-8
Date: Wed, 23 Mar 2016 00:30:26 GMT

0709ef7b70254ae074a42a558ee3b9de
```

## Increment and get the counter value
```
$ http PUT http://localhost:3000/0709ef7b70254ae074a42a558ee3b9de

HTTP/1.1 200 OK
Content-Length: 1
Content-Type: text/plain; charset=utf-8
Date: Wed, 23 Mar 2016 00:31:25 GMT

1
```
