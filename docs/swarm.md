# HOWTO run IronFunction as a scheduler on top of Docker Swarm cluster

*Prerequisite 1: Make sure you have a working Docker 1.12+ Swarm cluster in place, you can build one by following the instructions at [Docker's website](https://docs.docker.com/swarm/).*

*Prerequisite 2: It assumes that your running environment is already configured to use Swarm's master scheduler.*

This is a step-by-step procedure to execute IronFunction on top of Docker Swarm cluster. It works by having IronFunction daemon started through Swarm's master, and there working as a job scheduler distributing tasks through Swarm pipes. Most likely, Swarm's master will place the daemon

### Steps

1. Start IronFunction in the Swarm Master. It expects all basic Docker environment variables to be present (DOCKER_TLS_VERIFY, DOCKER_HOST, DOCKER_CERT_PATH, DOCKER_MACHINE_NAME). The important part is that the working Swarm master environment must be passed to Functions daemon:
```ShellSession
$ docker login # if you plan to use private images
$ docker volume create --name functions-datafiles
$ docker run -d --name functions \
        -p 8080:8080 \
        -e DOCKER_TLS_VERIFY \
        -e DOCKER_HOST \
        -e DOCKER_CERT_PATH="/docker-cert" \
        -e DOCKER_MACHINE_NAME \
        -v $DOCKER_CERT_PATH:/docker-cert \
        -v functions-datafiles:/app/data \
        iron/functions
```

2. Once the daemon is started, check where it is listening for connections:

```ShellSession
# docker info
CONTAINER ID        IMAGE                COMMAND                  CREATED             STATUS              PORTS                                     NAMES
5a0846e6a025        iron/functions       "/usr/local/bin/entry"   59 seconds ago      Up 58 seconds       2375/tcp, 10.0.0.1:8080->8080/tcp   swarm-agent-00/functions
````

Note `10.0.0.1:8080` in `PORTS` column, this is where the service is listening. IronFunction will use Docker Swarm scheduler to deliver tasks to all nodes present in the cluster.

3. Test the cluster:

```ShellSession
$ export IRON_FUNCTION=$(docker port functions | cut -d ' ' -f3)

$ curl -H "Content-Type: application/json" -X POST -d '{ "app": { "name":"myapp" } }' http://$IRON_FUNCTION/v1/apps
{"message":"App successfully created","app":{"name":"myapp","config":null}}

$ curl -H "Content-Type: application/json" -X POST -d '{ "route": { "type": "sync", "path":"/hello-sync", "image":"iron/hello" } }' http://$IRON_FUNCTION/v1/apps/myapp/routes
{"message":"Route successfully created","route":{"appname":"myapp","path":"/hello-sync","image":"iron/hello","memory":128,"type":"sync","config":null}}

$ curl -H "Content-Type: application/json" -X POST -d '{ "name":"Johnny" }' http://$IRON_FUNCTION/r/myapp/hello-sync
Hello Johnny!
```
