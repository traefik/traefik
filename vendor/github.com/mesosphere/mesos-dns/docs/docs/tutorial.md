---
title: Mesos-DNS Set Up
---

## Mesos-DNS on a Mesos Cluster

This is a step-by-step tutorial for running Mesos-DNS on a Mesos cluster.

### Step 1: Launch a Mesos cluster

Follow the instructions on the Mesosphere Community Documentation site to [set up a Mesos and Marathon cluster](https://open.mesosphere.com/getting-started/install/).

Have the IP addresses of your cluster to hand, since we'll be installing and running Mesos DNS by sshing into your nodes. Make sure that port `53` is unblocked for both `tcp` and `udp` traffic between your nodes. To unblock traffic for port `53` on Google Compute Engine, follow these [directions](http://stackoverflow.com/questions/21065922/how-to-open-a-specific-port-such-as-9090-in-google-compute-engine).

### Step 2: Install Mesos-DNS

We will install Mesos-DNS on the slave node `2.3.4.5` (replace this with your actual IP address). You can access the node through ssh using:

```
ssh 2.3.4.5
```

Fetch the binary for your system from GitHub. The full list of releases can be seen [here](https://github.com/mesosphere/mesos-dns/releases):

```
mkdir /usr/local/mesos-dns/
wget -O /usr/local/mesos-dns/mesos-dns https://github.com/mesosphere/mesos-dns/releases/download/v0.5.1/mesos-dns-v0.5.1-linux-amd64
```

In the same directory (`/usr/local/mesos-dns`), create a file named `config.json` with the following contents:

```
$ cat /usr/local/mesos-dns/config.json
{
  "zk": "zk://10.41.40.151:2181/mesos",
  "refreshSeconds": 60,
  "ttl": 60,
  "domain": "mesos",
  "port": 53,
  "resolvers": ["169.254.169.254","10.0.0.1"],
  "timeout": 5,
  "email": "root.mesos-dns.mesos"
}
```
The `resolvers` field includes the two nameservers listed in the `/etc/resolv.conf` of the nodes in this cluster.

### Step 3: Launch Mesos-DNS

We will launch Mesos-DNS from the master node
`1.2.3.4`. You can access the node through ssh using:

```
ssh 1.2.3.4
```

We will use Marathon to launch Mesos-DNS in order to get fault-tolerance. If Mesos-DNS crashes, Marathon will automatically restart it. Note that we constrain this to the node where we compiled Mesos-DNS earlier so that the binary is available. If this node goes away, however, then we will lose the Mesos-DNS service. In a production deployment, it is highly advisable to run multiple instances of Mesos-DNS.

Create a file `mesos-dns.json` with the following contents:

```
$ more mesos-dns.json 
{
"cmd": "sudo  /usr/local/mesos-dns/mesos-dns -v -config=/usr/local/mesos-dns/config.json",
"cpus": 1.0, 
"mem": 1024,
"id": "mesos-dns",
"instances": 1,
"constraints": [["hostname", "CLUSTER", "2.3.4.5"]]
}

```

Launch Mesos-DNS via Marathon using:

```
curl -X POST -H "Content-Type: application/json" http://1.2.3.4:8080/v2/apps -d@mesos-dns.json
```

This command instructs Marathon to launch Mesos-DNS on node `2.3.4.5`. You can access the `stdout` and `stderr` for Mesos-DNS through the Mesos webUI, accessible through `http://1.2.3.4:5050` in this example.

### Step 4: Configure cluster nodes

Next, we will configure all nodes in our cluster to use Mesos-DNS as their DNS server. Access each node through ssh and execute:


```
sudo sed -i '1s/^/nameserver 2.3.4.5\n /' /etc/resolv.conf
```

We can verify that the configuration is correct and that Mesos-DNS can server DNS queries using the following commands:

```
$ cat /etc/resolv.conf
nameserver 2.3.4.5
 domain c.myproject.internal.
search c.myprojecct.internal. 267449633760.google.internal. google.internal.
nameserver 169.254.169.254
nameserver 10.0.0.1
$ host www.google.com
www.google.com has address 74.125.70.104
www.google.com has address 74.125.70.147
www.google.com has address 74.125.70.99
www.google.com has address 74.125.70.105
www.google.com has address 74.125.70.106
www.google.com has address 74.125.70.103
www.google.com has IPv6 address 2607:f8b0:4001:c02::93
```

To be 100% sure that Mesos-DNS is actually the server that provided the translation above, we can try:

```
$ sudo apt-get install dnsutils
$ dig www.google.com

; <<>> DiG 9.8.4-rpz2+rl005.12-P1 <<>> www.google.com
;; global options: +cmd
;; Got answer:
;; ->>HEADER<<- opcode: QUERY, status: NOERROR, id: 45045
;; flags: qr rd ra; QUERY: 1, ANSWER: 6, AUTHORITY: 0, ADDITIONAL: 0

;; QUESTION SECTION:
;www.google.com.			IN	A

;; ANSWER SECTION:
www.google.com.		228	IN	A	74.125.201.105
www.google.com.		228	IN	A	74.125.201.103
www.google.com.		228	IN	A	74.125.201.147
www.google.com.		228	IN	A	74.125.201.104
www.google.com.		228	IN	A	74.125.201.106
www.google.com.		228	IN	A	74.125.201.99

;; Query time: 3 msec
;; SERVER: 2.3.4.5#53(2.3.4.5)
;; WHEN: Sat Jan 24 01:03:38 2015
;; MSG SIZE  rcvd: 212
```

The line marked `SERVER` makes it clear that the process we launched to listen to port `53` on node `2.3.4.5` is providing the answer. This is Mesos-DNS.

### Step 5: Launch nginx using Mesos

Now let's launch a task using Mesos. We will use the nginx webserver using Marathon and Docker. We will use the master node for this:


```
ssh 1.2.3.4
```

First, create a configuration file for nginx named `nginx.json`:

```
$ cat nginx.json
{
  "id": "nginx",
  "container": {
    "type": "DOCKER",
    "docker": {
      "image": "nginx:1.7.7",
      "network": "HOST"
    }
  },
  "instances": 1,
  "cpus": 1,
  "mem": 640,
  "constraints": [
    [
      "hostname",
      "UNIQUE"
    ]
  ]
}
```

You can launch it on Mesos using:

```
curl -X POST -H "Content-Type: application/json" http://1.2.3.4:8080/v2/apps -d@nginx.json
```

This will launch nginx on one of the three slaves using docker and host networking. You can use the Marathon webUI to verify it is running without problems. It turns out that Mesos launched it on node `3.4.5.6` and we can verify it works using:

```
$ curl http://3.4.5.6
<!DOCTYPE html>
<html>
<head>
<title>Welcome to nginx!</title>
<style>
    body {
        width: 35em;
        margin: 0 auto;
        font-family: Tahoma, Verdana, Arial, sans-serif;
    }
</style>
</head>
<body>
<h1>Welcome to nginx!</h1>
<p>If you see this page, the nginx web server is successfully installed and
working. Further configuration is required.</p>

<p>For online documentation and support please refer to
<a href="http://nginx.org/">nginx.org</a>.<br/>
Commercial support is available at
<a href="http://nginx.com/">nginx.com</a>.</p>

<p><em>Thank you for using nginx.</em></p>
</body>
</html>
```

### Step 6: Use Mesos-DNS to connect to nginx

Now, let's use Mesos-DNS to communicate with nginx. We will still use the master node:

```
ssh 1.2.3.4
```

First, let's do a DNS lookup for nginx, using the expected name `nginx.marathon.mesos`:

```
$ dig nginx.marathon.mesos

; <<>> DiG 9.8.4-rpz2+rl005.12-P1 <<>> nginx.marathon.mesos
;; global options: +cmd
;; Got answer:
;; ->>HEADER<<- opcode: QUERY, status: NOERROR, id: 11742
;; flags: qr aa rd ra; QUERY: 1, ANSWER: 1, AUTHORITY: 0, ADDITIONAL: 0

;; QUESTION SECTION:
;nginx.marathon.mesos. IN	A

;; ANSWER SECTION:
nginx.marathon.mesos. 60 IN	A	10.114.227.92

;; Query time: 0 msec
;; SERVER: 10.14.245.208#53(10.14.245.208)
;; WHEN: Sat Jan 24 01:11:46 2015
;; MSG SIZE  rcvd: 96

```

Mesos-DNS informed us that nginx is running on node `3.4.5.6`. Now let's try to connect with it:

```
$ curl http://nginx.marathon.mesos
<!DOCTYPE html>
<html>
<head>
<title>Welcome to nginx!</title>
<style>
    body {
        width: 35em;
        margin: 0 auto;
        font-family: Tahoma, Verdana, Arial, sans-serif;
    }
</style>
</head>
<body>
<h1>Welcome to nginx!</h1>
<p>If you see this page, the nginx web server is successfully installed and
working. Further configuration is required.</p>

<p>For online documentation and support please refer to
<a href="http://nginx.org/">nginx.org</a>.<br/>
Commercial support is available at
<a href="http://nginx.com/">nginx.com</a>.</p>

<p><em>Thank you for using nginx.</em></p>
</body>
</html>
```

We successfully connected with nginx using a logical name. Mesos-DNS works!


### Step 7: Scaling out nginx

Use the Marathon webUI to scale nginx to two instances. Alternatively, relaunch it after editing the json file in step 5 to indicate 2 instances. A minute later, we can look it up again using Mesos-DNS and get:

```
$  dig nginx.marathon.mesos

; <<>> DiG 9.8.4-rpz2+rl005.12-P1 <<>> nginx.marathon.mesos
;; global options: +cmd
;; Got answer:
;; ->>HEADER<<- opcode: QUERY, status: NOERROR, id: 30550
;; flags: qr aa rd ra; QUERY: 1, ANSWER: 2, AUTHORITY: 0, ADDITIONAL: 0

;; QUESTION SECTION:
;nginx.marathon.mesos. IN	A

;; ANSWER SECTION:
nginx.marathon.mesos. 60 IN	A	4.5.6.7
nginx.marathon.mesos. 60 IN	A	3.4.5.6

;; Query time: 1 msec
;; SERVER: 2.3.4.5#53(2.3.4.5)
;; WHEN: Sat Jan 24 01:24:07 2015
;; MSG SIZE  rcvd: 143
```

Now, Mesos-DNS is giving us two A records for the same name, identifying both instances of nginx on  our cluster. 



