# proc_exporter

A prometheus metrics exporter for your linux /proc folder.  


# about

Warning: This only works on linux-based systems and needs read privilege for the full /proc folder.

This is a simple http server which exports the "proc_cpu_time_total" metric.  
It does so by going through all the process IDs inside the proc directory, reading the stat file inside and adding the user and kernel cpu values together.  
The metric is named "cpu_time" but in reality it is the "amount of time that this process has been scheduled in user+kernel mode, measured in clock ticks".  
Source: [https://man7.org/linux/man-pages/man5/proc_pid_stat.5.html](https://man7.org/linux/man-pages/man5/proc_pid_stat.5.html)


# install

The release binary is statically compiled for 64bit linux systems.  
Download it and start it like this:

```bash
wget https://github.com/OverlordAkise/proc_exporter/releases/latest/download/proc_exporter-amd64.zip
unzip proc_exporter-amd64.zip
./proc_exporter
```


# run as service

The following systemd service file can be used to run the proc_exporter as a background service:

```
[Unit]
Description=proc_exporter for grafana/prometheus
After=network.target iptables.service

[Service]
Type=simple
WorkingDirectory=/srv/proc_exporter
ExecStart=/srv/proc_exporter/proc_exporter -local
Restart=always
RestartSec=1m
IPAccounting=yes
IPAddressDeny=any
IPAddressAllow=127.0.0.0/8

CapabilityBoundingSet=
RestrictAddressFamilies=~AF_PACKET
RestrictNamespaces=true
ProtectClock=true
ProtectControlGroups=true
ProtectHome=true
ProtectKernelLogs=true
ProtectKernelModules=true
ProtectKernelTunables=true
#ProtectProc=noaccess
ProtectSystem=full
SystemCallFilter=~@clock @debug @module @mount @reboot @swap @cpu-emulation @obsolete
LockPersonality=true
RemoveIPC=true
UMask=0027
RestrictRealtime=true
NoNewPrivileges=true
PrivateTmp=true
PrivateMounts=true
PrivateDevices=true
ProtectHostname=true
#ProcSubset=pid
RestrictSUIDSGID=true
PrivateUsers=true

InaccessiblePaths=/boot /media /mnt /opt /root /var
StandardOutput=append:/srv/proc_exporter/stdout.log
StandardError=inherit

[Install]
WantedBy=multi-user.target
```

In the above file update the "ExecStart" and "WorkingDirectory" and save the file as "/etc/systemd/system/proc_exporter.service".

To then start it run:

```bash
sudo systemctl daemon-reload
sudo systemctl start proc_exporter
```
