## integration-test

## Checklist
* Ensure you have Gatling distribution in the base dir (run `make install-deps`)
* Ensure you have GATLING_SERVER_INCOMING_S3_URL env var. This should point to your test s3 bucket.
* Ensure you have GATLING_SERVER_RESULTS_S3_URL env var. This should point to your test s3 bucket.
* Ensure you have AWS credentials setup correctly.

## Running

In the base dir: 
```
make test-all
```
