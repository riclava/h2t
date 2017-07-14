# h2t
  h2t is a program to transport tcp over http

## API usage
- GET all rules
```bash
curl 'http://localhost:8081/api'
```
```json
{
    "acl":{
        "172.23.27.87:22":{
            "name":"172.23.27.87:22",
            "date":"2017-07-15",
            "description":"server87"
        },
        "172.23.27.88:22":{
            "name":"172.23.27.88:22",
            "date":"2017-07-15",
            "description":"server88"
        }
    }
}
```
- PUT a new rule, if exists, will update it
```bash
curl -H "Content-Type: application/json" -X POST -d '{"name":"172.23.27.89:22","date":"2017-07-15","description":"server89"}' 'http://localhost:8081/api'
```
```
no content
```
- DELETE a single rule
```bash
curl -X DELETE 'http://localhost:8081/api/172.23.27.89:22'
```
```
no content
```
- DELETE all rules
```bash
curl -X DELETE 'http://localhost:8081/api'
```
```
no content
```
- Store all changes by PUT
```bash
curl -X PUT 'http://localhost:8081/api'
```
```
no content
```

## HTTP to TCP usage
- Test connect to server
```
curl -XCONNECT http://127.0.0.1:8081/
```
- Using inline
```
ssh root@172.23.27.87 -o "ProxyCommand=nc -X connect -x 127.0.0.1:8081 %h %p"
```
- Using ssh_config or ~/.ssh/config
```
# filter
Host host-which-need-proxy.com
  ProxyCommand nc -X connect -x 127.0.0.1:8081 %h %p
```
```
# global
Host *
   ProxyCommand nc -X connect -x 127.0.0.1:8081 %h %p
```

## tips
- no openbsd nc ? copy one from ubuntu, I've test that ubuntu 16.04.2's nc is valid
  - copy /bin/nc.openbsd to dst machine
  - ldd /bin/nc.openbsd
  - copy library of libbsd.so.0 to dst machine and make a soft link
  - ldconfig
  - done ;-)