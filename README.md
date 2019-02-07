# Description

ACE SDK allows developers to implement microservices that can be used as executable step in a choreography.

# Service Implemenation using the SDK
## Implement the callback method

### The callback method signature
Implement the callback method that will process the received message. Below is the signature of the method

```
    func businessMessageProcessor(ctx context.Context, businessMsg *messaging.BusinessMessage, msgProducer linker.MsgProducer) error
```

### Input processing

The businessMsg parameter of type messaging.BusinessMessage identifies the business message and holds the payload and the metadata associated with the payload. 

#### Processing payload

The payload can be retrieived using the interface method GetPayload() on businessMsg object. 

The payload [type messaging.Payload] can be actual payload (use messaging.Payload.GetBody() to retrieve the content) or a reference to a location (use messaging.Payload.GetLocationReference() to identify if the payload body is location reference)

#### Processing metadata

The metadata can be retrieved using the interface method GetMetaData() on businessMsg object which returns map of string key/value pairs identifying the metadata associated with the payload.

### Output processing
The output of the processing can be responded by generating new message(s). To create a new message construct business message and setup metadata. The new message can then be produced using clientRelay parameter. 

#### Creating new ACE business message
- Construct the metadata

    Constructing the metadata is done by setting up a map of string key/value pairs

- Contruct the payload

    Create the payload with content as demonstrated below. The newContent in the example below is a byte array holding the raw content
    ```
            newPayload := messaging.Payload{Body: newContent}
    ```

    OR

    Create the payload with location reference as demonstrated below. The newContent in the example below is a byte array holding the the location reference.
    ```
            newPayload := messaging.Payload{Body: newContent, LocationReference: true}
    ```

- Construct the ACE business message

    Create new business message object as demonstrated below. The "newMetadata" and "newPayload" in the example below identifies the metadata and payload for the new business message
    ```
        newBusinessMsg := messaging.BusinessMessage{ MetaData: newMetadata, Payload: &newPayload}
    ```

#### Producing message
To produce messages use the Send method on msgProducer parameter as demostrated below
```
    msgProducer.Send(ctx, &newBusinessMsg)
```

## Add trace for service execution (Optional)
ACE SDK has instrumentation for OpenTracing(https://opentracing.io/specification/) built-in and provides ability to allow the business service to inject the tracing spans.

To start a span as child of span managed by ACE, use StartSpanFromContext method from opentracing package. Using the created span the business service can log details as demonstrated below.

```
span, ctxWithSpan := opentracing.StartSpanFromContext(ctx, "business-service")
span.LogFields(opentractingLog.String("event", "processed message"))
span.Finish()
```

Note: When creating a new span, make sure to return the new created context back while producing the messages(if any) thru SDK using the Send method as demonstrated below
```
    msgProducer.Send(ctxWithSpan, &newBusinessMsg)
```

## Register the service callback method with ACE
ACE business service must register the service info and callback method for making it usable as a step to build choreographies
The service registration needs following details
- Service Name
- Service Version
- Service Description
- Callback method 

Below is an example of Registering the service info & callback method and then starting the ACE processing
```
aceAgent, err := linker.Register(serviceName, serviceVersion, serviceDescription, businessMessageProcessor)

aceAgent.Start()
```

The provided template reads the serviceName, serviceVersion and serviceDescription from following environment variables respectively, but its the implemenation choice on how to setup these details.
    - SERVICE_NAME
    - SERVICE_VERSION
    - SERVICE_DESCRIPTION
