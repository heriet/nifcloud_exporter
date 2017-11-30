# NIFCLOUD exporter

Prometeus exporter for NIFCLOUD

## Usage

create config.yml
```
---

rdb:
  - name: myrdb    # Environment name
    region: east-1 # Region of API endpoint 
    accessKeyId: <YOUR ACCESS KEY>
    secretAccessKey: <YOUR SECRET ACCESS KEY>
    instances:
      - name: mydbname # DB instance name
```

```
$ ./nifcloud_exporter --config.file=/path/to/config.yml
```

## supported service

- NIFCLOUD RDB

## License

This project is distributed under the Apache License, Version 2.0, see LICENSE.txt.