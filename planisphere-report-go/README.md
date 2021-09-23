# Planisphere Report Go

## Installing

Make sure you have the devil-ops tap intstalled with:

```
brew tap devil-ops/devil-ops git@gitlab.oit.duke.edu:devil-ops/homebrew-devil-ops.git
```

Install using:

```
brew install planisphere-report
```

## About

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

## Expanding to new Platforms

We've tried to abstract everthing we can in to the `lookup.go` file. The more
specific platform info can be found in the platform specific files, such as
`linux.go`, `darwin.go`, etc. If you would like to add your own OS here and
flesh out the collection bits, we'd love a pull request!

A starting point for the platform `foo` could be something like the following in
`foo.go`:

```go
func ApplyPlatformDetections(l *Lookuper) error {
    // Do initializing bits here
    return nil
}
func setSerial(l *Lookuper) (string, error) {
	// TODO: Implement this
	return "", errors.New("Not yet implemented")
}

func setManufacturer(l *Lookuper) (string, error) {
	// TODO: Implement this
	return "", errors.New("Not yet implemented")
}

func setModel(l *Lookuper) (string, error) {
	// TODO: Implement this
	return "", errors.New("Not yet implemented")
}

func setDiskEncrypted(l *Lookuper) (bool, error) {
	// TODO: Implement this
	return false, errors.New("Not yet implemented")
}

func setMemory(l *Lookuper) (uint64, error) {
	// TODO: Implement this
	return 0, errors.New("Not yet implemented")
}

func setOSFamily(l *Lookuper) (string, error) {
	// TODO: Implement this
	return "", errors.New("Not yet implemented")
}

func setDeviceType(l *Lookuper) (string, error) {
	// TODO: Implement this
	return "", errors.New("Not yet implemented")
}

func setOSFullName(l *Lookuper) (string, error) {
	// TODO: Implement this
	return "", errors.New("Not yet implemented")
}

func setUsername(l *Lookuper) (string, error) {
	// TODO: Implement this
	return "", errors.New("Not yet implemented")
}

func setInstalledSoftware(l *Lookuper) ([][]string, error) {
	// TODO: Implement this
	return nil, errors.New("Not yet implemented")

}
```