# scout
Tool for dumping live http traffic or PCAP files as json data.

Features:
1. dumps http request and corresponding response as single json object. Each pair as single-line json.
2. unpacks gzipped requests and responses bodies.
3. it is pretty fast (multiple goroutines for parsing).
4. supports request-response pairs grouping by regexp.
5. file-rotation for dumps for long live capturing.

## Why
The main purpose of this tool is to provide easy way for debugging live HTTP traffic. 
Very often application logs are not detailed enough, PCAP files from tcpdump are big and 
you don't want to reconfigure or redeploy your reverse-proxy with traffic dump option enabled.

Line-by-line output as json makes it extremely easy to parse, grep and format with `jq` utility.

## How to use
The simplest way is to run via docker container (change `eth0` to your network interface name):
```
docker run --rm -ti --net host ontrif/scout -i eth0
```

Now you can see live traffic in your terminal:
```
curl http://ipecho.net/plain
123.123.123.123
```

Json output from scout (formatted with `jq`):
```json
{
  "dst": "146.255.36.1",
  "dst_port": "80",
  "req": {
    "body": "",
    "headers": {
      "Accept": [
        "*/*"
      ],
      "User-Agent": [
        "curl/7.60.0"
      ]
    },
    "host": "ipecho.net",
    "method": "GET",
    "url": "/plain"
  },
  "res": {
    "body": "123.123.123.123",
    "code": 200,
    "headers": {
      "Cache-Control": [
        "no-cache"
      ],
      "Content-Type": [
        "text/plain"
      ],
      "Date": [
        "Sun, 22 Jul 2018 17:08:58 GMT"
      ],
      "Expires": [
        "Mon, 26 Jul 1997 05:00:00 GMT"
      ],
      "Pragma": [
        "no-cache"
      ],
      "Server": [
        "Apache"
      ],
      "Vary": [
        "Accept-Encoding"
      ]
    },
    "status": "200 OK"
  },
  "src": "192.168.0.2",
  "src_port": "55866"
}
```

