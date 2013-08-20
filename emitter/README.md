loggregatorlib/emitter
==================

This is a GO library to emit messages to the loggregator.

Create a emitter with NewEmitter with the loggregator server hostname and port, source type, and an gosteno logger.

Call Emit on the emitter with the application GUID and message strings.

Valid source types are:

 	CLOUD_CONTROLLER
 	ROUTER
 	UAA
 	DEA
 	WARDEN_CONTAINER


Sample Workflow
    import "github.com/cloudfoundry/loggregatorlib/emitter"

    func main() {
        emitter, err := emitter.NewEmitter("localhost:38452", "ROUTER", gosteno.NewLogger("LoggregatorEmitter"))
        emitter.Emit(appId, message)
    }