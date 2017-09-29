# planisphere-report

planisphere-report is a simple python client to planisphere's [self reporting mechanism](https://planisphere.oit.duke.edu/help/self_report).

## Supported Distributions

* Fedora 26+
* RHEL/Scientific/CentOS 6
* RHEL/Scientific/CentOS 7
* Ubuntu 14
* Ubuntu 16
* Arch Linux

Accepting patches for any other distributions..

## Install

### 1. Get Self Report key

If the device is managed by a Support Group, go to the page for the Support
Group in planisphere and get the self report key for that Support Group.

If the device is self managed by you, go to planisphere, click the 'My Profile'
link at the top of the page and get your self report key.

Once you have the key, put it in `/etc/planisphere-report-key`

For example:

```
$ echo YOUR_KEY_HERE | sudo tee /etc/planisphere-report-key
```

### 2. Install the script

Copy planisphere-report to ```/usr/bin/planisphere-report``` (or wherever you want) and make sure its executable.

### 3. Setup a cron job

Add a cronjob to run regularly (ie once a day) to run the script.  This should
run as root as linux does not allow non-root users to access the device's
serial number.

For laptops, consider calling the script after acquiring a DHCP address so
planisphere will have an accurate list of external MAC addresses for the laptop.

Alternatively consult [systemd service/timer and DHCP trigger](systemd/) in this repository

### 4. Config file (optional)

If you want to further configure how ```planisphere-report``` runs, you can
create a config file as ```/etc/planisphere-report```

The config file is a simple ini file, everything in the file is optional.  If
there is a ```[config]``` section, it will change some basic operating parameters.
If there's a ```[overrides]``` section, it'll override some of the values detected
by ```planisphere-report``` or add in options that ```planisphere-report``` cann't
detect on its own.   See https://planisphere.oit.duke.edu/help/self_report for more
details on the values that can be set in overrides.

Example:

```
[config]
url = https://planisphere.oit.duke.edu/self_report
key-file = /etc/planisphere-report-key

[overrides]
username = sean
```
