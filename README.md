# s3web

Serve static websites from any S3 compatible object storage endpoints.

## Usage

You can use `docker-compose` to run `s3web`:

```yaml
version: "3"
services:
  s3web:
    image: ghcr.io/long2ice/s3web/s3web
    network_mode: host
    restart: always
    volumes:
      - ./config.yaml:/config.yaml
```

## Configuration

This is example of the configuration file:

```yaml
server:
  listen: 0.0.0.0:8080
  logTimezone: Asia/Shanghai
  logTimeFormat: '2006-01-02 15:04:05.000000'
  compressLevel: 0
s3:
  endpoint: localhost:9000
  secure: false
  accessKey: minio
  secretKey: minio123
  bucket: mybucket
  region: us-east-1
sites:
  - domains: 
      - localhost # domain name
    subFolder: / # sub folder
    spa: false # single page application
```

## Credits

- [s3www](https://github.com/harshavardhana/s3www), base project for `s3web` improvement.

## License

This project is licensed under the
[Apache-2.0](https://github.com/long2ice/s3web/blob/master/LICENSE)
License.