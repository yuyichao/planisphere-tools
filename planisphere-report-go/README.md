# Planisphere Report Go

## Installing

### macOS

Make sure you have the devil-ops tap intstalled with:

```bash
brew tap devil-ops/devil-ops https://gitlab.oit.duke.edu/devil-ops/homebrew-devil-ops.git
```

Install using:

```bash
brew install planisphere-report
```

### Linux Packages

Download the appropriate package from the releases page and install. These
packages all install a systemd timer, set to run once a day

### Other

Download the appropriate binary from the releases page

Put the `planisphere-report` binary in an appropriate path, such as `/usr/local/bin`, and run:

```bash
planisphere-report install cron
```

This will set up a cronjob to run at a random hour and minute daily.

## About

Do some minimal setup by adding a `~/.planisphere-report.yaml`

Configuration can also be added to `/etc/planisphere-report.[extension]`

For backwards compatibility, you can also add your key as a text string in `/etc/planisphere-report-key`

These can also be set with ENV vars using a prefix of `PLANISPHEREREPORT_`, example:

```bash
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
    os_family: "macOS"
    memory_mb: 50000
    username: Joe User
    department_key: abcd
    status: deployed
    support_group_id: 5
    support_group_name: some_group
    usage_type: self_managed
    installed_software:
        - - "foo"
          - "1.2.3"
    extra_data:
        hello: world
```

## Expanding to new Platforms

Check out the files in `internal/os`. Create a new folder for your OS, as well
as some tests. To get going quickly, you can copy from an existing OS we already
support
