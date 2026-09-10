# Usage

## Submitting a Simulation

Simulations must be packaged as a self-contained (uber) jar containing the compiled simulation classes, resources,
and all dependencies — including Scala and Gatling itself — since the server runs it directly off that jar's
classpath. If you're using Maven, the [maven-shade-plugin](https://maven.apache.org/plugins/maven-shade-plugin/) can
build this for you; see [gatling-scala-example](https://github.com/jecklgamis/gatling-scala-example) for a working
setup that produces `target/gatling-scala-example.jar`.

### Via HTTP Upload

```bash
curl -v \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -F "file=@target/gatling-scala-example.jar" \
  -F "simulation=gatling.test.example.simulation.ExampleSimulation" \
  -F "javaOpts=-DbaseUrl=http://localhost:8080 -DdurationMin=1 -DrequestPerSecond=10" \
  http://localhost:58080/task/upload
```

The response includes a `taskId`, used to query the server for artifacts such as console logs or Gatling reports.

### Via a Generic Submit (upload once, submit anywhere)

Upload a jar to get a URL back, then submit a task referencing any http(s) or s3 URL — including the one you just
got:

```bash
curl -H "Authorization: Bearer ${API_TOKEN}" -F "file=@target/gatling-scala-example.jar" http://localhost:58080/upload
# => {"id":"<uuid>"}
```

Uploaded files are stored at `uploads/<uuid>/<filename>` and served directly (HTTP Basic Auth, see the
[API Reference](api.md)) from `/uploads/<uuid>/<filename>`.

```bash
curl -v -H "Content-Type: application/json" http://localhost:58080/task/submit -d @request.json
```

`request.json`:

```json
{
  "url": "http://localhost:58080/uploads/<uuid>/gatling-scala-example.jar",
  "simulation": "gatling.test.example.simulation.ExampleSimulation",
  "javaOpts": "-DbaseUrl=http://localhost:8080 -DdurationMin=0.10 -DrequestPerSecond=1"
}
```

`url` also accepts `s3://...` locations. This requires the S3 downloader to be enabled *and* scoped to specific
bucket(s) in `configs/config-<env>.yaml` — s3 downloads are rejected by default until `allowedBuckets` is set,
so a valid API token can't be used to read arbitrary buckets your AWS credentials can reach:

```yaml
downloaders:
  s3:
    enabled: true
    configMap:
      region: some-region
      allowedBuckets: gatling-server-incoming,another-bucket
```

Similarly, `url` for http(s) downloads is restricted to `localhost`/`127.0.0.1`/`::1` (so the self-referential
`/uploads` flow above keeps working) plus anything that resolves to a public IP — private, loopback, and link-local
addresses (including the cloud metadata endpoint) are rejected unless the host is explicitly added to
`taskSubmit.allowedHttpHosts` in `configs/config-<env>.yaml`:

```yaml
taskSubmit:
  allowedHttpHosts:
    - host: localhost
    - host: some.internal.host
      auth:
        type: bearer
        token: some-token
    - host: another.internal.host
      auth:
        type: basic
        username: some-user
        password: some-password
```

`auth` is optional per host and supports `basic` (`username`/`password`) or `bearer` (`token`); credentials are only
ever sent to their own host, never to others on the list or to a public-IP fallback match. A host with no `auth`
(like `localhost` above) gets the `browseAuth` credentials attached instead, which is what makes the self-referential
`/uploads` flow keep working now that `/uploads/` itself requires Basic Auth.

### Aborting a Task

```bash
curl -X POST -H "Authorization: Bearer ${API_TOKEN}" http://localhost:58080/task/abort/{taskId}
```

## Retrieving Artifacts

A simulation run produces a console log, Gatling report, simulation log, and the original request's metadata (see
the `/task/*` rows in the [API Reference](api.md)). These are available directly from the server, and are also
uploaded to S3 if an S3 uploader is configured. The test report is a downloadable `tar.gz` archive. The whole
workspace directory (one subdirectory per task, containing the raw files above) is also browsable directly at
`http://localhost:58080/workspace/{taskId}/` (also requires the bearer token).

## Authoring Simulations

Gatling simulations can be written in Scala, Java, or Kotlin. Simple simulations can be submitted as-is; for
anything more involved, a build project (Maven, for example) makes packaging much easier. See the example projects
for a working setup in your language of choice:

* [gatling-scala-example](https://github.com/jecklgamis/gatling-scala-example)
* [gatling-java-example](https://github.com/jecklgamis/gatling-java-example)
* [gatling-kotlin-example](https://github.com/jecklgamis/gatling-kotlin-example)
