# Pronsy: a DNS to DNS-over-TLS Proxy for N26 SRE Challenge

Pronsy is a DNS proxy written in Go that listen UDP and TCP requests from the
client and resolves the petitions against a DNS/TLS server, like CloudFlare or
Google.  

## Index
- [Run It](#Test-it-yourself!)
- [Configuration](#Configuration)
- [About My Implementation](#About-my-implementation)
    - [Design](#The-Design)
    - [UDP and TCP handlers](#UDP-and-TCP-concurrent-handlers-with-Bonus-Features)  
        - [Testing the UDP resolution](#Testing-the-UDP-resolution)
    - [Cache](#Cache---Bonus-Feature)
    - [Denylist and API](#Denylist-with-REST-API---Bonus-Feature)
    - [Logger](#Logger---Bonus-Feature) 
- [Challenge Questions](#Challenge-Questions)
## Test it yourself! 

```sh
## build and run it using docker
make docker-build && make docker-run

## run it in your pc
make run

## test it solving a domain
dig blog.charlei.xyz @127.0.0.1 -p 5353 +tcp
dig blog.charlei.xyz @127.0.0.1 -p 5353
```
## Configuration
All the configurations so far are made via environment variables. 
It is possible to set your own configurations from the following files
depending on how the application is launched. 

- In the `docker-compose.yaml` in case you are running it with `make docker-run` 
- In the `env.env` if running with `make run`

The variables in the files are almost self-explanatory, but they are also
mentioned along this document. 

## About my implementation
### The Design
A little speak about my code rather than the project itself. I wrote my code
using a blend of [Domain Driven
Design](https://martinfowler.com/bliki/DomainDrivenDesign.html) and [Clean
Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html).

From my perspective this is very useful for two main reasons: 
- Way easier to test compared to other patterns since it allows to create your
  own mocks and stubs with the use of interfaces and dependency injection
  (besides this project has no tests because of time reasons).

- The order. The domain is the most important and you shouldn't contamine it
  with implementation details. That's the job of the gateways and the
  controllers.
                                
### UDP and TCP concurrent handlers with Bonus Features   

Pronsy handles UDP and TCP DNS petitions. For the TCP implementation it uses an
atomic counter to limit the number of active connections and make use of
goroutines to handle concurrent requests. This can be configured with the
environment variable `PRONSY_TCPMAXCONNPOOL`

The UDP implementation has a more elaborated approach, it features a custom
queue created on top of a channel and the limit of the 'handled messages' is
set by the channel buffer size. The message to be solved the 'Proxy Service'
goes through a goroutine that reads the message from the socket and send it to
the queue. On the other side, there is a 'Dequeue' function that also works
concurrently getting the messages from the queue and sending them to the
handler function where the petition is finally solved calling the Proxy
Service. For the UDP implementation the goroutines are limited by the number of
CPUs the host machine have. 

It is possible to change the buffer size of the message queue with
`PRONSY_UDPMAXQUEUESIZE`. 

#### Testing the UDP resolution. 
I ran some tests under different conditions to see how the UDP resolution
behaves. All the tests were performed in my local machine, a Laptop with an
Intel i7-1165g7 8 cores @ 4.70ghz 16GB of RAM DDR4, using ArchLinux
5.16.1-arch1-1.

The tool used for the tests was
[DNSBlast](https://github.com/jedisct1/dnsblast). It sends randon domains to a
given DNS Resolver. 

```bash
$ time ./dnsblast 127.0.0.1 1000 100 5353
```

- Conditions: Using CloudFlare, buffer size of 1000, sending 1000 requests. 100
  request per second.  

```sh
Queries Sent: 1000
Queries Received: 1000
Elapsed time: 49.986s
Reply Rate: 20 pps
```

- Conditions: Using CloudFlare, buffer size of 10000, sending 1000 requests.
  100 request per second.  

```sh
Queries Sent: 1000
Queries Received: 1000
Elapsed time: 35.370s
Reply Rate: 28 pps
```

- Conditions: Using CloudFlare, buffer size of 100000, sending 1000 requests.
  100 request per second.  

```sh
Queries Sent: 1000
Queries Received: 1000
Elapsed time: 19.171s
Reply Rate: 52 pps
```

- Conditions: Using CloudFlare, buffer size of 1000000, sending 1000 requests.
  100 request per second.  

```sh
Queries Sent: 1000
Queries Received: 1000
Elapsed time: 20.011s
Reply Rate: 49 pps
```
It seems I found a limit here since 100000 or 1000000 buffer size perform
almost the same. 

I did the test directly against CloudFlare, no proxy in the middle, and this is
what I got:

```sh
Queries Sent: 1000
Queries Received: 631
Elapsed time: 12.941s
Reply Rate: 61 pps
```
As expected, CloudFlare performs better than my local proxy but also seems to
limit the requests you send them and I guess that's why I'm not receiving
response to all of the queries DNSBlast sent. 

More of this tests in are in TODO: put file. 

### The Resolver
The resolver, at a software development level, is the package that knows how to
talk with a DNS/TLS provider to solve domains. It hides the implementation
details to the domain.  

It automatically gets the TLS connection working and retrieves the RootCAs
needed, and that enables Pronsy to talk with different providers. 

By default it's using CloudFlare as DNS Provider. It can be changed
when the application is started changing the value of the `PRONSY_PROVIDERHOST`
environment variable. 

### Cache - Bonus Feature
Pronsy features a really basic 'home-made' in-memory cache that saves the
recently solved domains to avoid losing time querying against the DNS
Provider. 

It's just a map protected with a sync/Mutex that is locked and unlocked by the
goroutines accesing it. 

This feature can be disabled by setting the `PRONSY_CACHEENABLED` environment variable to
`false`. The data from the cache is flushed every N seconds. It's possible to
assign a value to that N with the environment variable `PRONSY_CACHETTL`.  

This cache implementation is not tied to the application and can be changed
easily if desired. All what is needed is to write a new implementation
compliant with the `proxy.Cache` interface. 

If I get to deploy this solution in a more 'production ready' environment I
would create an implementation to use Redis. That way I could spin multiple
replicas of Pronsy while having a centralized cache server shared among the
replicas. 

### Denylist with REST API - Bonus Feature
This feature was not fully developed because of time reasons. 

The domain package can be found with the interfaces needed to start this
service. Also a not fully developed Rest API to manage the denylist with
methods like `AddDomain`, `GetDomain`, `DeleteDomain` or `GetDomains` it's in
the codebase but it's not being used. The API is initialized but only answer to
`GET /ping`. 

The use case for this feature was to tell Pronsy to not resolve some blocked
domains by an administrator just like [PiHole](https://pi-hole.net/) or
[Blocky](https://0xerr0r.github.io/blocky/) do.  


### Logger - Bonus Feature
Most of the packages of Pronsy can be injected with a Logger. Just like the
cache, this can be changed for a different implementation as well as it
implements the methods of the Logger interface. 

The STD Output implementation shipped with this codebase allows Pronsy to log
messages in three different levels: 
- Debug
- Info
- Error

A useful implementation could be the integration with a 3rd party log service
such as AWS CloudWatch or any other custom service. It is possible (and
desirable, from my perspective) to write all the code and inject the dependency
to the packages to start loging to CloudWatch without modifying the domain of
our application. 
                                        
If I were to give a solution for a production environment running multiple
replicas of Pronsy I would use a std output scrapper that sends the logs to a
different system where me or a group of teams can watch them. Solutions like
Logstash/Kibana or Loki/Promtail can be really useful to accomplish this.    

## Challenge Questions

### Imagine this proxy being deployed in an infrastructure. What would be the security concerns you would raise? 

If somebody get to sniff the incoming traffic to the proxy they can get to know
the domains that are being queried by the clients since the traffic between the
client and the proxy travels without encryption.

For a correct use of this solution, it should be deployed in a private network
were only the clients you want to use it can reach it. That way the unencrypted
traffic only goes through your network and never reaches internet. The other
side of the proxy is more secure since the traffic travels encrypted to the
DNS/TLS Provider. 

### How would you integrate that solution in a distributed, microservices-oriented and containerized architecture?

I have, at least, two approaches, with its trade offs and concerns.  

One of them is to deploy it as a service with all the replicas needed
behind an internal load balancer and make it available for other services within a
private network. In AWS it can be configured as a DNS Server for a VPC
enforcing all the hosts of that network to use it. 

I have some concerns about the resolution of private domains, but I think it's
possible to make that a feature in Pronsy. 

On the other side, If you are using something like Kubernetes you can just
deploy Pronsy as a sidecar for every application in the cluster [the way Istio
does](https://istio.io/latest/docs/ops/configuration/traffic-management/dns-proxy/). 

I found two problems to solve with this approach.
- The observability. It's a complex task to get all the logs of all your
  sidecars in a big infrasctructure. Also, there is a pro about this approach
  regarding the observability: It could be easier to identify which
  pod/applications is doing which request. (Also a feature opportunity for
  Pronsy.) 
- The compute resources. 1 pod to 1 sidecar of DNS proxy can be an overkill most
  of the times and it's possible that having so many replicas of the proxy can
  require a bigger infrasctructure when it can be avoided. 

### What other improvements do you think would be interesting to add to the project?
I would go deeper with the development of the bonus features I added to the project. 
Logs feature can go further. I would create a log implementation that push logs away.
For a production environment the cache could be a key feature. 
The deny/block list it's a really nice to have here. This capability can extend to block domains or also block IPs. 
Metrics exposure. We already have logs. But we can add some metrics endpoints, Prometheus style, to create dashboards to quickly see how many petitions are solved successfully, how many of them failed, why they failed, the most petitioned domains, cache metrics, blocked domains metrics and son on. 


