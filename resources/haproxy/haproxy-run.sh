docker run -d --name my-haproxy \
    -v $(pwd)/haproxy.cfg:/usr/local/etc/haproxy/haproxy.cfg \
    -v /var/run/haproxy:/var/run/haproxy \
    -p 80:80 \
    haproxy:latest
