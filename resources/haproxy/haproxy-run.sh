#!/bin/bash

# Create the haproxy-run directory if it doesn't exist
if [ ! -d "./haproxy-run" ]; then
    mkdir haproxy-run
fi

# Copy the contents of the resources/haproxy directory to the haproxy-run directory
cp -r ./resources/haproxy/* ./haproxy-run

# Run the Docker container with HAProxy
docker run -d --name my-haproxy \
    --network=host \
    -v "${PWD}/haproxy-run:/usr/local/etc/haproxy:rw" \
    haproxytech/haproxy-ubuntu