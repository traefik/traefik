## Summary
These scripts were developed to deploy mesos-dns to a Mesos cluster created from the tooling at https://google.mesosphere/com.

## Usage
Edit the `hosts` file, replacing the addresses with those of your Mesos cluster, then:
```shell
$ ansible-playbook -i ./hosts site.yml
```
