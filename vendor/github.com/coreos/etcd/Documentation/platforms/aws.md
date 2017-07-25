## Introduction

This guide assumes operational knowledge of Amazon Web Services (AWS), specifically Amazon Elastic Compute Cloud (EC2). This guide provides an introduction to design considerations when designing an etcd deployment on AWS EC2 and how AWS specific features may be utilized in that context.

## Capacity planning

As a critical building block for distributed systems it is crucial to perform adequate capacity planning in order to support the intended cluster workload. As a highly available and strongly consistent data store increasing the number of nodes in an etcd cluster will generally affect performance adversely. This makes sense intuitively, as more nodes means more members for the leader to coordinate state across. The most direct way to increase throughput and decrease latency of an etcd cluster is allocate more disk I/O, network I/O, CPU, and memory to cluster members. In the event it is impossible to temporarily divert incoming requests to the cluster, scaling the EC2 instances which comprise the etcd cluster members one at a time may improve performance. It is, however, best to avoid bottlenecks through capacity planning.

The etcd team has produced a [hardware recommendation guide]( ../op-guide/hardware.md) which is very useful for “ballparking” how many nodes and what instance type are necessary for a cluster.

AWS provides a service for creating groups of EC2 instances which are dynamically sized to match load on the instances. Using an Auto Scaling Group ([ASG](http://docs.aws.amazon.com/autoscaling/latest/userguide/AutoScalingGroup.html)) to dynamically scale an etcd cluster is not recommended for several reasons including:

* etcd performance is generally inversely proportional to the number of members in a cluster due to the synchronous replication which provides strong consistency of data stored in etcd
* the operational complexity of adding [lifecycle hooks](http://docs.aws.amazon.com/autoscaling/latest/userguide/lifecycle-hooks.html) to properly add and remove members from an etcd cluster by modifying the [runtime configuration](../op-guide/runtime-configuration.md)

Auto Scaling Groups do provide a number of benefits besides cluster scaling which include:

* distribution of EC2 instances across Availability Zones (AZs)
* EC2 instance fail over across AZs
* consolidated monitoring and life cycle control of instances within an ASG

The use of an ASG to create a [self healing etcd cluster](#self-healing) is one of the design considerations when deploying an etcd cluster to AWS.

## Cluster design

The purpose of this section is to provide foundational guidance for deploying etcd on AWS. The discussion will be framed by the following three critical design criteria about the etcd cluster itself:

* block device provider: limited to the tradeoffs between EBS or instance storage (InstanceStore)
* cluster topology: how many nodes should make up an etcd cluster; should these nodes be distributed over multiple AZs
* managing etcd members: creating a static cluster of EC2 instances or using an ASG.

The intended cluster workload should dictate the cluster design. A configuration store for microservices may require different design considerations than a distributed lock service, a secrets store, or a Kubernetes control plane. Cluster design tradeoffs include considerations such as:

* availability
* data durability after member failure
* performance/throughput
* self healing

### Availability

Instance availability on AWS is ultimately determined by the Amazon EC2 Region Service Level Agreement ([SLA](https://aws.amazon.com/ec2/sla/)) which is the policy by which Amazon describes their precise definition of a regional outage.

In the context of an etcd cluster this means a cluster must contain a minimum of three members where EC2 instances are spread across at least two AZs in order for an etcd cluster to be considered highly available at a Regional level.

For most use cases the additional latency associated with a cluster spanning across Availability Zones will introduce a negligible performance impact.

Availability considerations apply to all components of an application; if the application which accesses the etcd cluster will only be deployed to a single Availability Zone it may not make sense to make the etcd cluster highly available across zones.

### Data durability after member failure

A highly available etcd cluster is resilient to member loss, however, it is important to consider data durability in the event of disaster when designing an etcd deployment. Deploying etcd on AWS supports multiple mechanisms for data durability.

* replication: etcd replicates all data to all members of the etcd cluster. Therefore, given more members in the cluster and more independent failure domains, the less likely that data stored in an etcd cluster will be permanently lost in the event of disaster.
* Point in time etcd snapshotting: the etcd v3 API introduced support for snapshotting clusters. The operation is cheap enough (completing in the order of minutes) to run quite frequently and the resulting archives can be archived in a storage service like Amazon Simple Storage Service (S3).
* Amazon Elastic Block Storage (EBS): an EBS volume is a replicated network attached block device which have stronger storage safety guarantees than InstanceStore which has a life cycle associated with the life cycle of the attached EC2 instance. The life cycle of an EBS volume is not necessarily tied to an EC2 instance and can be detached and snapshotted independently which means that a single node etcd cluster backed by an EBS volume can provide a fairly reasonable level of data durability.

### Performance/Throughput

The performance of an etcd cluster is roughly quantifiable through latency and throughput metrics which are primarily affected by disk and network performance. Detailed performance planning information is provided in the [performance section](../op-guide/performance.md) of the etcd operations guide.

#### Network

AWS offers EC2 Placement Groups which allow the collocation of EC2 instances within a single Availability Zone which can be utilized in order to minimize network latency between etcd members in the cluster. It is important to remember that collocation of etcd nodes within a single AZ will provide weaker fault tolerance than distributing members across multiple AZs. [Enhanced networking for EC2 instances](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/enhanced-networking.html) may also improve network performance of individual EC2 instances.

#### Disk

AWS provides two basic types of block storage: [EBS volumes](https://aws.amazon.com/ebs/) and [EC2 Instance Store](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/InstanceStorage.html). As mentioned, an EBS volume is a network attached block device while instance storage is directly attached to the hypervisor of the EC2 host. EBS volumes will generally have higher latency, lower throughput, and greater performance variance than Instance Store volumes. If performance, rather than data safety, is the primary concern it is highly recommended that instance storage on the EC2 instances be utilized. Remember that the amount of available instance storage varies by EC2 [instance types](https://aws.amazon.com/ec2/instance-types/) which may impose additional performance considerations.

Inconsistent EBS volume performance can introduce etcd cluster instability. [Provisioned IOPS](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/EBSVolumeTypes.html#EBSVolumeTypes_piops) can provide more consistent performance than general purpose SSD EBS volumes. More information about EBS volume performance is available [from AWS](https://aws.amazon.com/ebs/details/) and Datadog has shared their experience with [getting optimal performance with AWS EBS Provisioned IOPS](https://www.datadoghq.com/blog/aws-ebs-provisioned-iops-getting-optimal-performance/) in their engineering blog.

### Self healing

While using an ASG to scale the size of an etcd cluster is not recommended, an ASG can be used effectively to maintain the desired number of nodes in the event of node failure. The maintenance of a stable number of etcd nodes will provide the etcd cluster with a measure of self healing.

### Next steps

The operational life cycle of an etcd cluster can be greatly simplified through the use of the etcd-operator. The open source etcd operator is a Kubernetes control plane operator which deploys and manages etcd clusters atop Kubernetes. While still in its early stages the etcd-operator already offers periodic backups to S3, detection and replacement of failed nodes, and automated disaster recovery from backups in the event of permanent quorum loss.
