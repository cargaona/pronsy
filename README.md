## Pronsy: a DNS to DNS-over-TLS Proxy for N26 SRE Challenge

Pronsy is a DNS proxy written in Go that listen UDP and TCP requests from the
client and resolves the petitions against a DNS/TLS server, like CloudFlare or
Google.  

Test it yourself! 

```sh
## build and run it using docker
make docker-build && make docker-run

## run it in your pc
make run

## test it solving a domain
dig blog.charlei.xyz @127.0.0.1 -p 5353 +tcp
dig blog.charlei.xyz @127.0.0.1 -p 5353
```
By default it's using CloudFlare as DNS Provider. The provider can be changed
when the application is started changing the value of the `PRONSY_PROVIDERHOST`
environment variable. (In the `docker-compose.yaml` in case you are running it
with `make docker-run` or in the `env.env` if running with `make run`)

Pronsy automatically gets the TLS connection working and retrieving the RootCAs
needed. 

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
                                
### UDP and TCP concurrent handlers. (Bonus Feature) 
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

I ran some tests under different conditions to see how the UDP resolution
behaves. All the tests were performed in a Intel i7-1165g7 8 cores @ 4.70ghz
and 16GB of RAM DDR4. 

The tool used for the tests was
[DNSBlast](https://github.com/jedisct1/dnsblast).

```bash
$ time ./dnsblast 127.0.0.1 1000 100 5354
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
what I got

```sh
Queries Sent: 1000
Queries Received: 631
Elapsed time: 12.941s
Reply Rate: 61 pps

```
As expected, CloudFlare performs better than my local proxy but also seems to
limit the request you send them and I guess that's why I'm not receiving
response to all of the queries DNSBlast sent. 

More of this tests in are in TODO: put file. 


### Cache (Bonus Feature)
Pronsy features a really basic 'home-made' in-memory cache that saves the
recently solved domains to avoid losing time querying against the DNS
Provider. 

It's just a map protected with a sync/Mutex and is locked and unlocked by the
goroutines accesing to them. 

It can be disabled by setting the `PRONSY_CACHEENABLED`environment variable to
`false`

This cache implementation is not tied to the application and can be changed
easily if desired. All what is needed is to write a new implementation
compliant with the `proxy.Cache` interface. 

If I get to deploy this solution in a more 'production ready' environment I
would create an implementation to use Redis. That way I could spin multiple
replicas of Pronsy while having a centralized cache server. 

### Denylist with REST API (Bonus Feature)
This feature was not fully developed because of time reasons. 

The domain package can be found with the interfaces needed to start this
service. Also a not fully developed Rest API to manage the denylist with
methods like `AddDomain`, `GetDomain`, `DeleteDomain` or `GetDomains` it's in
the codebase but it's not being used. The API is initialized but only answer to
`GET /ping`. 

The use case for this feature was to tell Pronsy to not resolve some blocked
domains by an administrator just like [PiHole](https://pi-hole.net/) or
[Blocky](https://0xerr0r.github.io/blocky/) do.  


### Logger (Bonus Feature)
Most of the packages of Pronsy can be injected with a Logger. Just like the
cache, this is can be changed for a different implementation as well as it
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
replicas of Pronsy I would use a stdoutput scrapper that sends the logs to a
different system where me or a group of teams can watch them. Solutions like
Logstash/Kibana or Loki/Promtail can be really useful to accomplish this.    

## Challenge Questions

### Imagine this proxy being deployed in an infrastructure. What would be the security concerns you would raise?


### How would you integrate that solution in a distributed, microservices-oriented and containerized architecture?
