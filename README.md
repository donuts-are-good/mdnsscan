# mdnsscan

## usage

here's how you use mdnsscan:
```
mdnsscan
```

it will start looking for mdns service entries and scanning the ports of the devices found. mdnsscan will attempt to identify the service associated with each open port.

note: the current version of the application will scan ports 1 through 10000 on each device found. it may take a while depending on the number of devices and their response times.

### example output
here is an example output from mdnsscan:

```
found entry:
name:  test device
host:  test-device.local
addrv4:  192.168.1.10
addrv6:  ::1
port:  5000
info:  map[device:test device]
-------------------
scanning ports...
scanning port 1/10000...
scanning port 2/10000...
...
scanning port 22/10000...
port 22 is open
service message: ssh-2.0-openssh_7.9
port 22 is likely associated with ssh
...
scanning port 80/10000...
port 80 is open
service message: http/1.1 200 ok
port 80 is likely associated with http
got http response from http://192.168.1.10:80: 200 ok
beginning of response body from http://192.168.1.10:80: <!doctype html>...
...
scanning port 10000/10000...
total devices: 1
total open ports: 2
service ssh found 1 times
service http found 1 times
```
in this example, mdnsscan found a single device with two open ports: 22 and 80.


## license

MIT License 2023 donuts-are-good, for more info see license.md
