# Opendog

A hacky translation layer that converts Datadog traces into OpenTelemetry Protobufs for use with Jaegar

It's not perfect and there is a bunch of stuff to be implemented and tidied up but it's good enough to pipe Datadog traces locally into [Jaegar All In One](https://www.jaegertracing.io/docs/1.39/getting-started/#all-in-one)

It's also a successor to my very lightweight but handy tool [Spanner](https://github.com/marcus-crane/spanner) which I use to sanity check a lot of APM instrumentation without the overhead of running an entire Datadog agent.

## Usage

```console
$ go run .
2022/12/10 23:01:45 Opendog is listening on 127.0.0.1:8126
2022/12/10 23:01:46 Succesfully converted and forwarded 2 spans
2022/12/10 23:01:46 Succesfully converted and forwarded 2 spans
```

Your Datadog APM client will automatically connect if it is running on localhost.

If you'd like to point a container at it using Docker, you can set the following environment variable against your container:
* Windows/macOS: `DD_AGENT_HOST=host.docker.internal`
* Linux: `DD_AGENT_HOST=172.17.0.1`

Eventually it'll be packaged up properly, provide a docker container etc etc

## Example of Datadog traces piped into Jaegar

![](./docs/example.png)