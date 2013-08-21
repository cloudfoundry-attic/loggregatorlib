loggregatorlib
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

   go get ./...
   go test -v ./...