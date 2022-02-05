# Proxy UDP/TCP DNS to DoT (DNS over TLS)

This proxy server listen UDP and TCP traffic and redirect the request to a DNS over TLS resolver.

It has two ways of working. The first (Direct method) redirects the traffic from one connection to another (only available for TCP incoming connections) and the other (Not-Direct method) parses the DNS message before sending it to the resolver.
If the "Not-Direct" method is set (Check how on the *Configuration* section) further management of the DNS message can be made. Like loggin, caching etc..

## Configuration :wrench:

Configuration parameters are set in a single file, weather you are running on your host :computer: or docker :whale:. 

```sh
# .env file
PROXY_CONFIG_UDP_PORT=4545 # Port in which the proxy will listen to incomming UDP connections.
PROXY_CONFIG_TCP_PORT=4545 # Port in which the proxy will listen to incomming TCP connections. 
PROXY_CONFIG_METHOD=direct # (Optional) ('direct' for direct method or leave it blank for normal method)
PROXY_CONFIG_TCP_MAX_CONN_POOL=15 # Max TCP pool connections allowed
PROXY_CONFIG_CACHE_TTL=45 # The time the cache will hold a response in seconds. (Note: Cache is not implemented)
PROXY_RESOLVER_READ_TO=500 # Read timeout of the DoT resolver TLS connection in miliseconds.
```
## 💻 Build and Run on Host 

```sh
go build -o dns-proxy
source <(cat .env | awk '{print "export "$1}')
./dns-proxy
```
or simply run 
```sh
source <(cat .env | awk '{print "export "$1}')
go run main.go config.go # no need to build 
```

## 🐋 Build and Run with Docker
```sh
docker image build -t dns-tls-proxy .
docker container run --rm --env-file .env -p 4545:4545/tcp -p 4545:4545/udp proxy-dns
```

## Usage
```sh
dig lucascontre.site  @127.0.0.1 -p 4545 # UDP
dig lucascontre.site  @127.0.0.1 -p 4545 +tcp # TCP
```

# Resources 

* https://developers.cloudflare.com/1.1.1.1/dns-over-tls
* `openssl s_client -connect 1.1.1.1:853 -showcerts` - *to check cloudflare's root-certificate To check cloudflare's root-certificate.*
* https://gobyexample.com/atomic-counters - *for handling concurrent tcp contection-pool counter*
* [RFC 1035](https://www.ietf.org/rfc/rfc1035.txt) (Specially section 4.2.2)

---
# Questions
#### Imagine this proxy being deployed in an infrastructure. What would be the security concerns you would raise?

First of all the proxy receives unencrypted UDP and TCP traffic. So if this proxy is serving on a different infra than its clients that traffic may be sniffed. Same can happen locally but its more unlikely since the attacker should already have access to your machine.

#### How would you integrate that solution in a distributed, microservices-oriented and containerized architecture?

I see two possibles solution.
1. I can make this daemon run as a side-cars along with the apps that doesn't support DNS over TLS as a possible solutions to encrypting my DNS requests. Running on the host directly on the host is another option but deploy a change on the proxy might be a bit harder to make. 
2. Or I can expose it a separate microservice and use it as my DNS for all the other microservices. (This one would face the first security concern mentioned in the 1st question)

#### What other improvements do you think would be interesting to add to the project?
* :heavy_check_mark: Done
* (:heavy_plus_sign::heavy_minus_sign:) Partially Done
* :heavy_multiplication_x: Not Done

**Bonus:**
1. Allow multiple incoming requests at the same time (Bonus) :heavy_check_mark: 
2. Also handle UDP requests, while still querying tcp on the other side (Bonus) :heavy_check_mark:

**Aditional improvements:**

1. More tests ! (Just a small part of the domain is tested) `go test ./...` :heavy_multiplication_x:
2. A limited pool of connections to prevent a DDoS attack, since each tcp connection runs a new thread. :heavy_check_mark:
3. A cache that prevents querying the resolver if the petition has been made recently. :heavy_plus_sign: :heavy_minus_sign: ( Not implemented due to time reasons but yet used by the domain)
4. Run the solution on a [distroless](https://github.com/GoogleContainerTools/distroless) container to reduce attack surface.  :heavy_check_mark:
5. Move `config.go` to its own package :heavy_multiplication_x:
6. Play more with the parsed message `dnsmessage.Message`. Maybe add a blacklist, a logger of what's been requested. :heavy_multiplication_x:
