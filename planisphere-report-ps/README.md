# planisphere-report-ps

planisphere-report.ps1 is a simple PowerShell client to Planisphere's [self reporting mechanism](https://planisphere.oit.duke.edu/help/self_report).

## Supported Distributions

* Windows PowerShell
* Microsoft PowerShell running on Windows

Accepting patches for any other distributions.

## Install

### 1. Get Self Report key

If the device is managed by a Support Group, go to the page for the Support Group in planisphere and get the self report key for that Support Group.

If the device is self managed by you, go to planisphere, click the 'Logged in as NetID' link at the top of the page and get your self report key.



Once you have the key, you have several options:

+ Put it in ```/etc/planisphere-report-key```. For example:
    ```
    PS C:\> "YOUR_KEY_HERE" | Out-File /etc/planisphere-report-key
    ```
+ Put it in a planisphere-report settings file, either directly or as a path to a keyfile.
+ Pass it as an optional parameter, either directly or as a path to a keyfile.

### 2. Install the script

Copy planisphere-report.ps1 to ```/usr/local/bin/``` or ```C:\scripts\``` (or wherever you want).

### 3. Setup a Scheduled Task

Add a Scheduled Task to run regularly (at least once a day, but as often as once an hour) to run the script. This should run as SYSTEM as Windows does not allow non-Administrator users to access the device's encryption information, software installed for other users, or the Security Event Log.

For laptops, consider additionally calling the script after acquiring a DHCP address so that planisphere will have an accurate list of network information.

For more information, see [SCHEDULE](planisphere-report-ps/SCHEDULE.md) in this repository.

You can also view, download, modify (if necessary), and import a [Windows Task Scheduler XML file](planisphere-report-ps/planisphere-report-ps__hourly__DHCP_.xml) with the recommended settings already configured. This XML file expects th script to be stored at "C:\usr\local\bin\planisphere-report.ps1". If your script is stored anywhere else, or you have other arguments to pass to the script, you'll need to edit the task either before (as XML) or after (in Task Scheduler) importing it.

### 4. Config file (optional)

If you want to further configure how ```planisphere-report.ps1``` runs, you can create a config file as ```/etc/planisphere-report``` or pass a different location of a config file as a script parameter.

The config file is a simple ini file, everything in the file is optional. If there is a ```[config]``` section, it will change some basic operating parameters. If there's an ```[overrides]``` section, it'll override some of the values detected by ```planisphere-report.ps1``` or add in options that ```planisphere-report.ps1``` can't detect on its own. See https://planisphere.oit.duke.edu/help/self_report for more details on the values that can be set in overrides.

Example:

```
[config]
url = https://planisphere.oit.duke.edu/self_report
key-file = /etc/planisphere-report-key

[overrides]
username = jcstraff

[extra_data]
Location = Power House, 3rd Floor
```
