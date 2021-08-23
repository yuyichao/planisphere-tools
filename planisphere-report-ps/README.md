# planisphere-report-ps

planisphere-report-ps is a simple PowerShell client to planisphere's [self reporting mechanism](https://planisphere.oit.duke.edu/help/self_report).

## Supported Distributions

* Windows PowerShell
* Microsoft PowerShell running on Windwos

Accepting patches for any other distributions..

## Install

### 1. Get Self Report key

If the device is managed by a Support Group, go to the page for the Support Group in planisphere and get the self report key for that Support Group.

If the device is self managed by you, go to planisphere, click the 'Logged in as NetID' link at the top of the page and get your self report key.

Once you have the key, put it in `/etc/planisphere-report-key`

For example:

```
$ "YOUR_KEY_HERE" | Out-File /etc/planisphere-report-key
```

### 2. Install the script

Copy planisphere-report to ```/usr/local/bin/planisphere-report``` or ```C:\scripts\``` (or wherever you want).

### 3. Setup a Scheduled Task

Add a Scheduled Task to run regularly (at least once a day, but as often as once an hour) to run the script. This should
run as SYSTEM as Windows does not allow non-Administrator users to access the device's encryption information, software installed for other users, or the Security Event Log.

For laptops, consider calling the script after acquiring a DHCP address so
planisphere will have an accurate list of external MAC addresses for the laptop.

For more information, see [SCHEDULE](SCHEDULE.md) in this repository.

### 4. Config file (optional)

If you want to further configure how ```planisphere-report-ps``` runs, you can
create a config file as ```/etc/planisphere-report[.ini]```

The config file is a simple ini file, everything in the file is optional.  If
there is a ```[config]``` section, it will change some basic operating parameters.
If there's a ```[overrides]``` section, it'll override some of the values detected
by ```planisphere-report``` or add in options that ```planisphere-report``` can't
detect on its own.   See https://planisphere.oit.duke.edu/help/self_report for more
details on the values that can be set in overrides.

Example:

```
[config]
url = https://planisphere.oit.duke.edu/self_report
key-file = /etc/planisphere-report-key

[overrides]
username = jcstraff
```
