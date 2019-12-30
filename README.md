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

$ docker exec -i msgbox-mysql sh -c 'exec mysql -uroot' < ~/dev/workspace/go-projects/src/github.com/ghsbhatia/msgbox/dbsetup.sql

#### MongoDB

$ docker run --name msgbox-mongo  -e MONGO_INITDB_ROOT_USERNAME=root -e MONGO_INITDB_ROOT_PASSWORD=secret -d -p 8081:8081 -p 27017:27017  mongo:4.0-xenial

### User Admin Commands

Create a new user
```
$ curl -X POST -H "Content-Type: application/json" -d '{"username":"Bob"}' http://localhost:6060/users
```
Create a new group
```
$ curl -X POST -H "Content-Type: application/json" -d '{"groupname":"Engineering", "usernames":["Bob", "Doug", "Carol"]}'  http://localhost:6060/groups
```
Get group
```
$ curl -X GET http://localhost:6060/groups/Engineering
```
### Message Store Commands

Send message to User
```
$ curl -X POST -H "Content-Type: application/json" -d '{"sender":"Alice","recipient":{"username":"Bob"},"subject":"test","body":"test message"}'  #http://localhost:6080/messages
```
Reply to message 
```
$ curl -X POST -H "Content-Type: application/json" -d '{"sender":"Bob","subject":"re:test","body":"test message"}'  http://localhost:6080/messages/5e07aa4a4f40b77a6fa2b4ad/replies
```
Send message to group 
```
$ curl -X POST -H "Content-Type: application/json" -d '{"sender":"Alice","recipient":{"groupname":"Engineering"},"subject":"gtest","body":"group test message"}'  http://localhost:6080/messages
```
Get message
```
$ curl -X GET http://localhost:6080/messages/5e07fa1e76d55d43972e69f3
```
Get user messages
```
$ curl -X GET http://localhost:6080/users/Bob/mailbox
```
### Docker Image creation

$ docker build -t mboxuseradminsvc:1.0 . -f cmd/useradmin/Dockerfile

$ docker build -t mboxmsgstoresvc:1.0 . -f cmd/msgstore/Dockerfile

### Nginx Setup

Coming soon!

### Docker Compose 

Coming soon!
