# terraform-provider-clarity

Terraform provider for [clarity.st](https://clarity.st), see further [documentation](https://docs.clarity.st).


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


## Release

Trigger the `release` action by pushing a new tag.

```
git tag -a v0.0.1 -m "Added documentation"
git push origin v0.0.1
```
