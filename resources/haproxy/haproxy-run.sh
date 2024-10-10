if (-Not (Test-Path "./haproxy-run")) { mkdir haproxy-run }; cp -r ./resources/haproxy/* ./haproxy-run


docker run -d --name my-haproxy `
    -v "${PWD}/haproxy-run:/usr/local/etc/haproxy:rw" `
    -p 80:80 `
    -p 8404:8404 `
    -p 5555:5555 `
    haproxytech/haproxy-ubuntu

