# 1 million TCP connections

After watching the awesome talk
[Eran Yanay - 1 million websockets connections in Go](https://www.youtube.com/watch?v=LI1YTFMi8W4)
I decided to reproduce it for TCP connections (i.e. without the overhead for
websocket) with the code present in this repo.

The idea is to run one server instance on bare metal and multiple clients in
Docker where each opens 25.000 connections and sends dummy messages every 10s.
The reason for running the clients in Docker is that one TCP connection takes
one port and there is a limited amount ephermal ports. As each connection is
uniquely represented by a 5-tuple

```
(local-IP, local-port, remote-IP, remote-port, protocol)
```

we can simply run a Docker container with a fresh local IP to open as many
connections as we would like.

## Preparation

First, one needs to allow more file descriptors by appending

```
* soft nproc  10000000
* hard nproc  10000000
* soft nofile 10000000
* hard nofile 10000000
```

to `/etc/security/limits.conf` and

```
fs.file-max = 10000000
```

to `/etc/sysctl.conf` and then reboot

```sh
reboot
```

Before, running the server we also need to temporarily change `conntack_max` via

```sh
sudo sysctl -w net.nf_conntrack_max=2000000

# check with
sudo sysctl -a | grep nf_conntrack_max
```

## Running stuff

1st prepare server

```sh
go build ./cmd/server/
sudo ./server
```

then start clients

```sh
# preperation
docker build . -t tcp-client
docker network create my-net

# 40 instances are needed
docker run -d --network my-net tcp-client --host <ip of server> -m -n 25000
```

where you need to replace `<ip of server>` with the IP of the server as shown in
in the output of step 1 or `ip addr`.

## Verifying number of connections

The server shows the number of active TCP connections. Another way to see the
number of live connections is

```sh
sudo sysctl -a | grep nf_conntrack_count
```

## Results

I tested it on a Mini PC

```
Ubuntu 24.04 LTS
Intel N100
16 GB RAM
```

and was able to open even more than 1 million TCP connections. The limitig
factor was RAM as the Docker containers for the clients take up quite a lot of
space.

I started 10 client containers with 25.000 connections at once, waited till
all connections were established and then started the next 10 client containers.

On a more powerfull PC thinks should work faster.

**Remark:** I installed Docker as described
[here](https://docs.docker.com/engine/install/ubuntu/#installation-methods).
Using Docker Desktop for Linux might not do the job as it runs in VMs and is
much less performant.
