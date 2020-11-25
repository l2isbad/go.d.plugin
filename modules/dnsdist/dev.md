### Docker setup

This setup uses https://github.com/HSRNetwork/docker-dnsdist.

Ensure that the specified `DNSDIST_VERSION` [version is available](https://pkgs.alpinelinux.org/packages)
on your chosen `ALPINE_VERSION`.

-   Clone the repository

```
git clone https://github.com/HSRNetwork/docker-dnsdist
cd docker-dnsdist
```

-   Put the following config to the `dnsdist.conf.tmpl`

```
setACL({'0.0.0.0/0'})

newServer({address="1.1.1.1", qps=1000})
newServer({address="8.8.8.8", qps=1000})

webserver("0.0.0.0:8083", "pass", "key", {}, "0.0.0.0/0")
```

-   Build the docker image

```
ALPINE_VER="edge"
DNSDIST_VER="1.5.1-r2"

docker build \
  --build-arg DNSDIST_VERSION=$DNSDIST_VER \
  --build-arg ALPINE_VERSION=$ALPINE_VER \
  -t $(whoami)/dnsdist:$DNSDIST_VER .
```

-   Start `DNSdist` docker container

```
docker run -d -p 8083:8083 "$(whoami)/dnsdist:$DNSDIST_VER"
```

-   Verify the statistics

```
# using `apikey`
curl -H 'X-Api-Key: key' http://127.0.0.1:8083/jsonstat?command=stats

# using HTTP basic auth
curl -u#:pass http://127.0.0.1:8083/jsonstat?command=stats
```

### Installing from Packages/Source 

[Install DNSdist](https://dnsdist.org/install.html) on your environment, if you 
decide to compile the source code, it will be necessary to install the following
packages before:

- [colm](https://www.colm.net/open-source/colm/) programming language.
- [kelbt](http://freshmeat.sourceforge.net/projects/kelbt) parser.
- [ragel](https://www.colm.net/open-source/ragel/) state machine  compiler.

Create the configuration file, it is necessary to enable the webserver inside it:

```
newServer("8.8.8.8")
webserver("127.0.0.1:8083", "netdata", "netdata")
```

Start the server:

```cmd
dnsdist -C dnsdist.conf --local=0.0.0.0:5300
```

Do requests for the server:

```cmd
for a in {0..1000}; do dig netdata.cloud @127.0.0.1 -p 5300 +noall +nocookie > /dev/null; done
```

Verify the statistics:

```cmd
curl -H"X-API-Key: netdata" "http://127.0.0.1:8083/jsonstat?command=stats"
```