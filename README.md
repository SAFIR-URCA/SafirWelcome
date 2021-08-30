<p align="center">
  <img src="https://i.imgur.com/h7TH7bt.png" width="350" title="hover text">
  </p>
<h1 align="center">SafirWelcome</h1>


## Requirements

* go => 1.16

## Build / Setup

```bash
# Init
/sbin/useradd -s /usr/sbin/nologin -r safir_welcome
/bin/sudo /bin/mkdir /opt/safir_welcome 

# Build
/bin/go build 
/bin/sudo /bin/mv SafirWelcome  /opt/safir_welcome/

# Setup service 
/bin/sudo cp safir_welcome.service /opt/safir_welcome/
/bin/ln -s /opt/safir_welcome/safir_welcome.service /etc/systemd/system/safir_welcome.service
/bin/systemctl enable safir_welcome.service

# Set privileges
/bin/chown safir_welcome:safir_welcome /opt/safir_welcome -R
/bin/chmod 500 /opt/safir_welcome/SafirWelcome
/bin/chown root:root /opt/safir_welcome/safir_welcome.service
/bin/chmod 400 /opt/safir_welcome/safir_welcome.service
```

## Run / stop bot 

```bash
# Run service/bot
/bin/systemctl start safir_welcome.service

# Stop service/bot
/bin/systemctl stop safir_welcome.service
```
