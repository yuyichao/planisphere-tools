# Planisphere Report Go

Do some minimal setup by adding a `~/.planisphere-report.yaml`

```
---
key: foo
## Uncomment to use test
# url: https://planisphere-test.oit.duke.edu/self_report

## Overrides will always take precedence over auto-detected values
overrides:
  username: Joe User
```

These can also be set with ENV vars using a prefix of `PLANISPHEREREPORT_`, example:
```
export PLANISPHEREREPORT_KEY=foo
```
