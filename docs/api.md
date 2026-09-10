# API Reference

| Endpoint                         | Method | Auth   | Body / Params                                              | Description                                                     |
|-----------------------------------|--------|--------|--------------------------------------------------------------|-------------------------------------------------------------------|
| `/`                                | GET    | —      | —                                                            | Root info                                                          |
| `/buildInfo`                       | GET    | —      | —                                                            | Version/branch info                                                |
| `/probe/ready`                     | GET    | —      | —                                                            | Readiness probe                                                    |
| `/probe/live`                      | GET    | —      | —                                                            | Liveness probe                                                     |
| `/task/upload`                     | POST   | Bearer | multipart: `file`, `simulation`, `javaOpts`                   | Upload a jar and submit + run it in one call                       |
| `/upload`                          | POST   | Bearer | multipart: `file`                                            | Upload any file; returns `{"id": "<uuid>"}`                        |
| `/uploads/{id}/{filename}`         | GET    | Basic  | —                                                            | Download/browse an uploaded file                                   |
| `/task/submit`                     | POST   | Bearer | JSON: `simulation`, `javaOpts`, `url` (http(s) or s3)         | Download the jar from `url` and submit + run it                    |
| `/task/{taskId}`                   | GET    | Bearer | —                                                            | Task runtime status                                                 |
| `/task/metadata/{taskId}`          | GET    | Bearer | —                                                            | Original submission metadata                                        |
| `/task/console/{taskId}`           | GET    | Bearer | —                                                            | Raw JVM console log                                                 |
| `/task/simulationLog/{taskId}`     | GET    | Bearer | —                                                            | Gatling's own simulation log                                        |
| `/task/results/{taskId}`           | GET    | Bearer | —                                                            | Results archive (`results.tar.gz`)                                  |
| `/task/abort/{taskId}`             | POST   | Bearer | —                                                            | Kill a running task                                                 |
| `/workspace/{taskId}/...`          | GET    | Basic  | —                                                            | Browse raw task workspace files                                     |
| `/blackhole`                       | POST   | —      | —                                                            | No-op sink (default HTTP event-notifier target)                     |

`Bearer` means `Authorization: Bearer <API_TOKEN>` is required. `Basic` means HTTP Basic Auth is required, using
`browseAuth.username`/`browseAuth.password` from `configs/config-<env>.yaml` — or the `BROWSE_USERNAME`/
`BROWSE_PASSWORD` env vars, which take precedence when set. All three (`API_TOKEN`, `BROWSE_USERNAME`,
`BROWSE_PASSWORD`) default to `default` if left unset. Basic Auth is used for these two endpoints specifically
because they're meant to be browsed directly: the browser's native login prompt lets you click through file
listings without attaching a header by hand. Either scheme's missing/invalid credentials get `401 Unauthorized`,
and repeated failures from the same client are rate-limited with `429 Too Many Requests`.
