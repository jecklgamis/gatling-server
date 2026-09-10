# Deployment

## Using Docker

```bash
docker run -it --name gatling-server -p 58080:58080 -e API_TOKEN=some-secret-token jecklgamis/gatling-server:main
```

## Using a Prebuilt Binary

Download a release for your platform from [GitHub Releases](https://github.com/jecklgamis/gatling-server/releases), then:

```bash
tar xzf gatling-server-<os>-<arch>-<version>.tar.gz
cd gatling-server-<os>-<arch>-<version>
API_TOKEN=some-secret-token ./bin/run-server.sh
```

## Deploy to Kubernetes

A Helm chart is at [`deployment/k8s/helm`](https://github.com/jecklgamis/gatling-server/tree/main/deployment/k8s/helm):

```bash
cd deployment/k8s/helm
helm install gatling-server ./chart \
  --set apiToken=<token> \
  --set browseUsername=<user> \
  --set browsePassword=<password>
```

By default it deploys behind an nginx Ingress with cert-manager TLS at `gatling-server.jecklgamis.com`. `apiToken`
gates `/task/*`; `browseUsername`/`browsePassword` gate `/workspace/` and `/uploads/` (HTTP Basic Auth). All three
default to `"default"`, matching the app's own built-in fallback, so this works out of the box but should be
overridden for anything beyond a quick trial. Set `existingSecretName` instead to source all three from a Secret
you manage yourself, without putting them in `values.yaml`/git.

Set `persistence.enabled=true` to keep task workspaces (uploaded jars, console/simulation logs, results) across
pod restarts - task state is also kept in memory per-pod, so keep `replicaCount` at 1 if you enable it. See
`values.yaml` for all options.
