# msgbox

### Introduction

This repository provides a bare minimum implementation for a ficticious msgbox service. 
The implementation is based on go-kit library https://github.com/go-kit/kit for building microservices. 
Two microservices are provided - useradminservice to manage user admin functions so that new users and groups can be created, 
msgstoreservice to create and query messages as well as send replies to messages.

![alt text](https://iotmechanic.s3.us-east-2.amazonaws.com/code-sample/img/msgbox.jpg)

### Database Setup

#### MySQL

$ docker run --name msgbox-mysql -e MYSQL_USER=root -e MYSQL_ALLOW_EMPTY_PASSWORD=yes -e MYSQL_DATABASE=msgbox -d -p3306:3306 mysql:8.0

$ docker exec -i msgbox-mysql sh -c 'exec mysql -uroot' < $GOPATH/src/github.com/ghsbhatia/msgbox/dbsetup.sql

#### MongoDB

$ docker run --name msgbox-mongo  -e MONGO_INITDB_ROOT_USERNAME=root -e MONGO_INITDB_ROOT_PASSWORD=secret -d -p 8081:8081 -p 27017:27017  mongo:4.0-xenial

### Fetch vendor dependencies

msgbox $ glide update

### Unit Testing

```
$ go test  -v ./pkg/useradmin
$ go test  -v ./pkg/msgstore
```

### Build and Run Services

msgstore $ go build msgstoreservice.go
msgstore $ ./msgstoreservice

useradmin $ go build useradminservice.go
useradmin $ ./useradminservice

### User Admin Commands

Create users
```
$ curl -X POST -H "Content-Type: application/json" -d '{"username":"Bob"}' http://localhost:6060/users
$ curl -X POST -H "Content-Type: application/json" -d '{"username":"Doug"}' http://localhost:6060/users
$ curl -X POST -H "Content-Type: application/json" -d '{"username":"Carol"}' http://localhost:6060/users
$ curl -X POST -H "Content-Type: application/json" -d '{"username":"Alice"}' http://localhost:6060/users
```
Create a new group
```
$ curl -X POST -H "Content-Type: application/json" -d 
'{"groupname":"Engineering", "usernames":["Bob", "Doug", "Carol"]}'  http://localhost:6060/groups
```
Get user
```
$ curl -X GET http://localhost:6060/users/Bob
```
Get group
```
$ curl -X GET http://localhost:6060/groups/Engineering
```
### Message Store Commands

Send message to User
```
$ curl -X POST -H "Content-Type: application/json" -d 
'{"sender":"Alice","recipient":{"username":"Bob"},"subject":"test","body":"test message"}'  
http://localhost:6080/messages
```
Reply to message - Substitute <msgid> with id of message returned by previous command
```
$ curl -X POST -H "Content-Type: application/json" -d 
'{"sender":"Bob","subject":"re:test","body":"test message"}'  
http://localhost:6080/messages/<msgid>/replies
```
Send message to group 
```
$ curl -X POST -H "Content-Type: application/json" -d 
'{"sender":"Alice","recipient":{"groupname":"Engineering"},"subject":"gtest","body":"group message"}'  
http://localhost:6080/messages
```
Get message
```
$ curl -X GET http://localhost:6080/messages/<msgid>
```
Get user messages
```
$ curl -X GET http://localhost:6080/users/Bob/mailbox
```
Get replies
```
$ curl -X GET http://localhost:6080/messages/<msgid>/replies
```  
### Docker Image creation

$ docker build -t dhsbhatia/mboxuseradminsvc:1.0 . -f cmd/useradmin/Dockerfile

$ docker build -t dhsbhatia/mboxmsgstoresvc:1.0 . -f cmd/msgstore/Dockerfile

$ docker build -t dhsbhatia/mboxproxy:1.0 cmd/nginx -f cmd/nginx/Dockerfile

### Nginx Setup

Refer to nginx.conf for request passing rules. This configuration is specific to the docker-compose.yml.

### Docker Compose 

The images have already been pushed to dockerhub (https://hub.docker.com/u/dhsbhatia), so docker-compose.yml can be used without building the docker images locally.

#### Bring up stack

$ docker-compose --project-name mbox -f docker-compose.yml up 

#### Setup Database

$ docker exec -i useradmindb sh -c 'exec mysql -uroot' < $GOPATH/src/github.com/ghsbhatia/msgbox/dbsetup.sql

#### Execute Commands

User Admin and Message Store commands mentioned above can be executed against the msgbox stack running using dockerized containers. As the Nginx proxy is listening on port 80, all the requests can be sent to default http port and proxy will pass the request to correct service e.g. the following command shows how to get replies for a given message.

Get replies
```
$ curl -X GET http://localhost/messages/<msgid>/replies
```


