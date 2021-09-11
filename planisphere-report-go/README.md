# Planisphere Report Go

Do some minimal setup by adding a `~/.planisphere-report.yaml`

Configuration can also be added to `/etc/planisphere-report.[extension]`

For backwards compatibility, you can also add your key as a text string in `/etc/planisphere_key_file`

These can also be set with ENV vars using a prefix of `PLANISPHEREREPORT_`, example:
```
export PLANISPHEREREPORT_KEY=foo
```

```yaml
---
key: foo
## Uncomment to use test
# url: https://planisphere-test.oit.duke.edu/self_report

## Overrides will always take precedence over auto-detected values
## Here are some examples:
overrides:
    instance_key: "foo"
    os_family: "Fast and Furious"
    memory_mb: 50000
    username: Joe User
    department_key: heyo
    status: deployed
    support_group_id: 5
    support_group_name: ssi_systems
    usage_type: self_managed
    installed_software:
        - - "foo"
          - "1.2.3"
    extra_data:
        hello: world
```

