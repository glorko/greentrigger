docker run -d --name my-haproxy \
    -v $(pwd)/resources/haproxy/haproxy.cfg:/usr/local/etc/haproxy/haproxy.cfg \
    -p 80:80 \
    -p 8404:8404 \
    haproxy:latest

#