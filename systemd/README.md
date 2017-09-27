Systemd service and timer
===================================

Systemd unit files to run the planisphere report client

## Set up the unit files

Move `planisphere-report.service` and `planisphere-report.timer` to `/etc/systemd/system/`

Execute `systemctl daemon-reload` and then `systemctl enable --now planisphere-report.timer`. planisphere-report will be scheduled to run in 1 miutes, and subsequently it will run 24 hours from the prior run (if it cannot run at the scheduled time, it will run again as soon as the system is up)

## Set up the DHCP hooks

If you restart the daemon, the report will run 180 seconds later. You can use this behavior to trigger a run when e.g. DHCP state changes.

If you use networkmanager, you can accomplish this by copying `planisphere-timer-restart` to `/etc/NetworkManager/dispatcher.d`. The next time you get a lease, the service will be restarted and the report will be posted 1 minute later.
