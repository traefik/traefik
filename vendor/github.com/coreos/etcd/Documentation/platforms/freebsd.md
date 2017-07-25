# FreeBSD

Starting with version 0.1.2 both etcd and etcdctl have been ported to FreeBSD and can be installed either via packages or ports system. Their versions have been recently updated to 0.2.0 so now etcd and etcdctl can be enjoyed on FreeBSD 10.0 (RC4 as of now) and 9.x, where they have been tested. They might also work when installed from ports on earlier versions of FreeBSD, but it is untested; caveat emptor.

## Installation

### Using pkgng package system

1. If pkg­ng is not installed, install it with command `pkg` and answering 'Y' when asked.

2. Update the repository data with `pkg update`.

3. Install etcd with `pkg install coreos-etcd coreos-etcdctl`.

4. Verify successful installation by confirming `pkg info | grep etcd` matches:

```
r@fbsd­10:/ # pkg info | grep etcd
coreos­etcd­0.2.0              Highly­available key value store and service discovery
coreos­etcdctl­0.2.0           Simple commandline client for etcd
r@fbsd­10:/ #
```

5. etcd and etcdctl are ready to use! For more information about using pkgng, please see: http://www.freebsd.org/doc/handbook/pkgng­intro.html
 
### Using ports system

1. If ports is not installed, install with `portsnap fetch extract` (it may take some time depending on hardware and network connection).

2. Build etcd with `cd /usr/ports/devel/etcd && make install clean`. There will be an option to build and install documentation and etcdctl with it.

3. If etcd wasn't installed with etcdctl, it can be built later with `cd /usr/ports/devel/etcdctl && make install clean`.

4. Verify successful installation by confirming `pkg info | grep etcd` matches:
 

```
r@fbsd­10:/ # pkg info | grep etcd
coreos­etcd­0.2.0              Highly­available key value store and service discovery
coreos­etcdctl­0.2.0           Simple commandline client for etcd
r@fbsd­10:/ #
```

5. etcd and etcdctl are ready to use! For more information about using ports system, please see: https://www.freebsd.org/doc/handbook/ports­using.html

## Issues

If there are any issues with the build/install procedure or there's a problem that is local to FreeBSD only (for example, by not being able to reproduce it on any other platform, like OSX or Linux), please send a problem report using this page for more information: http://www.freebsd.org/send­pr.html
