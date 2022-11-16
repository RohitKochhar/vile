#!/bin/bash

curl http://localhost:8080/v1/key/key1 -X PUT -d "value1"
echo ""
curl http://localhost:8080/v1/key/key1 -X GET
echo ""
curl http://localhost:8080/v1/key/key1 -X PUT -d "value2"
echo ""
curl http://localhost:8080/v1/key/key1 -X GET
echo ""
curl http://localhost:8080/v1/key/key1 -X DELETE
echo ""
curl http://localhost:8080/v1/key/key1 -X GET
echo
curl http://localhost:8080/v1/key/key1 -X PUT -d "value1"
echo ""
curl http://localhost:8080/v1/key/key1 -X GET
