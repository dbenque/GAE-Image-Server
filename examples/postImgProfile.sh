#!/bin/sh
curl -w "\n" -X POST -d @- http://localhost:8080/imgProfile --header "Content-Type:application/json"

