---
title: Mesos-DNS using systemd
---

## Mesos-DNS using systemd

This is a step-by-step tutorial for running [Mesos-DNS](https://github.com/mesosphere/mesos-dns) with [systemd](http://freedesktop.org/wiki/Software/systemd/). 

### Step 1: Launch a Mesosphere cluster on a systemd platform

The first step is to create a Mesosphere cluster on a platform that supports systemd.   While all linux distros are moving in the direction of supporting systemd, the currently released list can be boiled down to RHEL 7, Suse v12 and CoreOs.  Mesosphere's DCOS is installed on CoreOS, however in this case Mesos-DNS is pre-installed.   This tutorial assumes that you are installing a fresh Mesosphere cluster without using Mesos-DNS.  For help installing please refer to [https://docs.mesosphere.com/getting-started/datacenter/install/](https://docs.mesosphere.com/getting-started/datacenter/install/).

This tutorial assumes the following cluster topology:

* One Master node which includes:
	* mesos-master
	* zookeeper
	* marathon
	* IP 
* Several slave nodes which have mesos-slave running on them.

For this tutorial we will assume the master IP address is `10.14.245.208`.

### Step 2: Install Mesos-DNS

We will install Mesos-DNS on node `10.14.245.208`.  Access the node through ssh.

After downloading a [release](https://github.com/mesosphere/mesos-dns/releases) and
extracting it, place the executable in well known location such as
`/usr/bin/mesos-dns`.

In the `/etc/mesos-dns/` directory lets create a file named `config.json` with the following contents: 

```
sudo mkdir /etc/mesos-dns/

$ cat /etc/mesos-dns/config.json 
{
  "zk": "zk://10.14.245.208:2181/mesos",
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

### Step 3: Running under systemd

In the `/etc/systemd/system` directory lets create a file named mesos-dns.service with the following contents:

```
[Unit]
Description=Mesos-DNS
After=network.target
Wants=network.target

[Service]
ExecStart=/usr/bin/mesos-dns -config=/etc/mesos-dns/config.json
Restart=on-failure
RestartSec=20

[Install]
WantedBy=multi-user.target
```

Now start the service:

`sudo systemctl start mesos-dns`

### Step 4: Configure cluster nodes

Next, we will configure all nodes in our cluster to use Mesos-DNS as their DNS server. Access each node through ssh and execute: 


```
sudo sed -i '1s/^/nameserver 10.14.245.208\n /' /etc/resolv.conf
```

We can verify that the configuration is correct and that Mesos-DNS can server DNS queries using the following commands:

```
$ cat /etc/resolv.conf 
nameserver 10.14.245.208
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
;; SERVER: 10.14.245.208#53(10.14.245.208)
;; WHEN: Sat Jan 24 01:03:38 2015
;; MSG SIZE  rcvd: 212
```

The line marked `SERVER` makes it clear that the process we launched to listen to port `53` on node `10.14.245.208` is providing the answer. This is Mesos-DNS. 

### Step 5: Launch nginx using Mesos

Now let's launch a task using Mesos. We will use the nginx webserver using Marathon and Docker. We will use the master node for this:

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
curl -X POST -H "Content-Type: application/json" http://10.41.40.151:8080/v2/apps -d@nginx.json
```

This will launch nginx on one of the three slaves using docker and host networking. You can use the Marathon webUI to verify it is running without problems. It turns out that Mesos launched it on node `10.114.227.92` and we can verify it works using:

```
$ curl http://10.114.227.92
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

First, let's do a DNS lookup for nginx, using the expected name `nginx.marathon-0.8.1.mesos`. The version number of Marathon is there because it registed with Mesos using name `marathon-0.8.1`. We could have avoided this by launching Marathon using ` --framework_name marathon`:

```
$ dig nginx.marathon-0.8.1.mesos

; <<>> DiG 9.8.4-rpz2+rl005.12-P1 <<>> nginx.marathon-0.8.1.mesos
;; global options: +cmd
;; Got answer:
;; ->>HEADER<<- opcode: QUERY, status: NOERROR, id: 11742
;; flags: qr aa rd ra; QUERY: 1, ANSWER: 1, AUTHORITY: 0, ADDITIONAL: 0

;; QUESTION SECTION:
;nginx.marathon-0.8.1.mesos. IN	A

;; ANSWER SECTION:
nginx.marathon-0.8.1.mesos. 60 IN	A	10.114.227.92

;; Query time: 0 msec
;; SERVER: 10.14.245.208#53(10.14.245.208)
;; WHEN: Sat Jan 24 01:11:46 2015
;; MSG SIZE  rcvd: 96

```

Mesos-DNS informed us that nginx is running on node `10.114.227.92`. Now let's try to connect with it:

```
$ curl http://nginx.marathon-0.8.1.mesos
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
$  dig nginx.marathon-0.8.1.mesos

; <<>> DiG 9.8.4-rpz2+rl005.12-P1 <<>> nginx.marathon-0.8.1.mesos
;; global options: +cmd
;; Got answer:
;; ->>HEADER<<- opcode: QUERY, status: NOERROR, id: 30550
;; flags: qr aa rd ra; QUERY: 1, ANSWER: 2, AUTHORITY: 0, ADDITIONAL: 0

;; QUESTION SECTION:
;nginx.marathon-0.8.1.mesos. IN	A

;; ANSWER SECTION:
nginx.marathon-0.8.1.mesos. 60 IN	A	10.29.107.105
nginx.marathon-0.8.1.mesos. 60 IN	A	10.114.227.92

;; Query time: 1 msec
;; SERVER: 10.14.245.208#53(10.14.245.208)
;; WHEN: Sat Jan 24 01:24:07 2015
;; MSG SIZE  rcvd: 143
```

Now, Mesos-DNS is giving us two A records for the same name, identifying both instances of nginx on our cluster. 



