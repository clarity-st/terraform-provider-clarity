# Examples



### Local development

```
# ~/.terrform.rc

provider_installation {
  dev_overrides {
    "local/clarity" = "{ $HOME }/go/bin"
  }
}
```

NB: replace `{ $HOME }`


```
# export CLARITY_API_TOKEN=...

terraform plan

terraform apply

terraform destroy

```
