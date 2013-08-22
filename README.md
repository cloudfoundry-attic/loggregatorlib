loggregatorlib [![Build Status](https://travis-ci.org/cloudfoundry/loggregatorlib.png?branch=master)](https://travis-ci.org/cloudfoundry/loggregatorlib)
==================

Loggregatorlib includes packages and libraries used by loggregator.

The emitter is an external library used to emit messages to the loggregator server.


Setup
------------------

    export GOPATH=`pwd`

    mkdir -p src/github.com/cloudfoundry

    cd src/github.com/cloudfoundry

    git clone https://github.com/cloudfoundry/loggregatorlib.git



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
*   logtarget: LogTarget contains the id of an app that is the target of a logmessage
*   testhelpers: Helpers for testing
