loggregatorlib/emitter
==================

This is a GO library to emit messages to the loggregator.

Create a emitter with NewEmitter with the loggregator trafficcontroller hostname and port, source type, and an gosteno logger.

Call Emit on the emitter with the application GUID and message strings.

### Valid source types are:

 	CLOUD_CONTROLLER
 	ROUTER
 	UAA
 	DEA
 	WARDEN_CONTAINER

###Sample Workflow

    import "github.com/cloudfoundry/loggregatorlib/emitter"

    func main() {
        appGuid := "a8977cb6-3365-4be1-907e-0c878b3a4c6b" // The GUID(UUID) for the user's application
        emitter, err := emitter.NewEmitter("10.10.10.16:38452", "ROUTER", gosteno.NewLogger("LoggregatorEmitter"))
        emitter.Emit(appGuid, message)
    }

###TODO

* By default all messages are annotated with a message type of OUT.
* At this time, we don't support emitting messages with a message type of ERR.

