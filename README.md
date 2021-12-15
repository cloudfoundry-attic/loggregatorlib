#loggregatorlib 

[![Coverage Status](https://coveralls.io/repos/cloudfoundry/loggregatorlib/badge.png?branch=master)](https://coveralls.io/r/cloudfoundry/loggregatorlib?branch=master)

Loggregatorlib includes packages and libraries used by loggregator.

The emitter is an external library used to emit messages to the loggregator server.

DEPRECATED
------------------
This library and its contents are no longer in use by Loggregator components.
Packages that were still in use have been moved to the loggregator repository.

Setup
------------------

    export GOPATH=`pwd`

    go get github.com/cloudfoundry/loggregatorlib



Running Tests:
------------------

    cd loggregatorlib
    go get ./...
    go test -i --race ./...
    go test -v --race ./...
    
Conform to `go vet`
------------------
    go vet ./...


Components
------------------

*   cfcomponent: Components used by Loggregator for use with CloudFoundry.
*   emitter:  GO library to emit messages to the loggregator. For instructions see the emitter/README.
*   loggregatorclient: A package used to send UDP messages. Used by Emitter and DEAagent.
*   logmessage: The package for loggregator protobuffer messages.
*   appid: Contains the id of an app that is the target of a logmessage
*   lib_testhelpers: Helpers for testing
