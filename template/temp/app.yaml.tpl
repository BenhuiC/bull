base:
  env: dev

http:
  listenAddr: :8080

database:
  db: root:123456@tcp(localhost:3306)/dev?charset=utf8mb4&parseTime=True&loc=Local
  workerRedisURL: redis://:@localhost:6379/0
  cronRedisURL: redis://:@localhost:6379/0
  redisInstanceURL: redis://:@localhost:6379/0

ceph:
  endpoint: "https://ns01obs.gaccloud.com.cn"
  accessKeyID: "BDFD0C035FE0EF14024F"
  accessKeySecret: "7BGlU7+2GZ4JdQcs5SwLZQaw1CwAAAGDX+DvGNuk"
  bucket: "spe_data"
  disableEndpointHostPrefix: true
  s3ForcePathStyle: true
  insecureSkipVerify: true
  signHost: "https://ns01obs.gaccloud.com.cn"

hasher:
  salt: "123"

services:
  stub:
    addr: "http://localhost:18080"

elastic:
  host: "http://127.0.0.1:9200"
  username: "spe"
  password: "Spe2077@#$%"
