#!/bin/bash
set -x

./build.sh

# test it
docker stop test-postgres-func
docker rm test-postgres-func

docker run -p 5432:5432 --name test-postgres-func -d postgres
sleep 5s

docker run --rm -i --link test-postgres-func:postgres postgres psql -h postgres -U postgres -c 'CREATE TABLE people (first TEXT, last TEXT, age INTEGER);'

RECORD1='{
    "first": "John",
    "last": "Smith",
    "age": 30
}'
echo $RECORD1 | docker run --rm -i -e SERVER=postgres:5432 -e COMMAND=INSERT -e TABLE=people --link test-postgres-func:postgres iron/func-postgres

QUERY1='{
    "last": "Smith"
}'
echo $QUERY1 | docker run --rm -i -e SERVER=postgres:5432 -e COMMAND=SELECT -e TABLE=people --link test-postgres-func:postgres iron/func-postgres

RECORD2='{
    "first": "Bob",
    "last": "Smith",
    "age": 43
}'
echo $RECORD2 | docker run --rm -i -e SERVER=postgres:5432 -e COMMAND=INSERT -e TABLE=people --link test-postgres-func:postgres iron/func-postgres

echo $QUERY1 | docker run --rm -i -e SERVER=postgres:5432 -e COMMAND=SELECT -e TABLE=people --link test-postgres-func:postgres iron/func-postgres

QUERY2='{
    "first": "John",
    "last": "Smith"
}'
echo $QUERY2 | docker run --rm -i -e SERVER=postgres:5432 -e COMMAND=SELECT -e TABLE=people --link test-postgres-func:postgres iron/func-postgres

docker stop test-postgres-func
docker rm test-postgres-func