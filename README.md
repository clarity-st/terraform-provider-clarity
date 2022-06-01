# terraform-provider-clarity



## Testing

For installed provider testing, see [`examples/basic`](examples/basic/README.md)

For [acceptance testing](https://www.terraform.io/plugin/sdkv2/testing/acceptance-tests)

```
# export CLARITY_API_TOKEN=...

# Optionally override AWS settings (see internal/clarity_provider_test.go)
# export AWS_ACCOUNT_ID=...
# export AWS_REGION=...
# export CLARITY_PROVIDER_ROLE=...

TF_ACC=1 go test -v ./internal
```
