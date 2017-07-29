Release v1.6.18 (2017-01-27)
===

Service Client Updates
---
* `service/clouddirectory`: Adds new service
  * Amazon Cloud Directory is a highly scalable, high performance, multi-tenant directory service in the cloud. Its web-based directories make it easy for you to organize and manage application resources such as users, groups, locations, devices, policies, and the rich relationships between them.
* `service/codedeploy`: Updates service API, documentation, and paginators
  * This release of AWS CodeDeploy introduces support for blue/green deployments. In a blue/green deployment, the current set of instances in a deployment group is replaced by new instances that have the latest application revision installed on them. After traffic is rerouted behind a load balancer to the replacement instances, the original instances can be terminated automatically or kept running for other uses.
* `service/ec2`: Updates service API and documentation
  * Adds instance health check functionality to replace unhealthy EC2 Spot fleet instances with fresh ones.
* `service/rds`: Updates service API and documentation
  * Snapshot Engine Version Upgrade

Release v1.6.17 (2017-01-25)
===

Service Client Updates
---
* `service/elbv2`: Updates service API, documentation, and paginators
  * Application Load Balancers now support native Internet Protocol version 6 (IPv6) in an Amazon Virtual Private Cloud (VPC). With this ability, clients can now connect to the Application Load Balancer in a dual-stack mode via either IPv4 or IPv6.
* `service/rds`: Updates service API and documentation
  * Cross Region Read Replica Copying (CreateDBInstanceReadReplica)

Release v1.6.16 (2017-01-24)
===

Service Client Updates
---
* `service/codebuild`: Updates service documentation and paginators
  * Documentation updates
* `service/codecommit`: Updates service API, documentation, and paginators
  * AWS CodeCommit now includes the option to view the differences between a commit and its parent commit from within the console. You can view the differences inline (Unified view) or side by side (Split view). To view information about the differences between a commit and something other than its parent, you can use the AWS CLI and the get-differences and get-blob commands, or you can use the GetDifferences and GetBlob APIs.
* `service/ecs`: Updates service API and documentation
  * Amazon ECS now supports a state for container instances that can be used to drain a container instance in preparation for maintenance or cluster scale down.

Release v1.6.15 (2017-01-20)
===

Service Client Updates
---
* `service/acm`: Updates service API, documentation, and paginators
  * Update for AWS Certificate Manager: Updated response elements for DescribeCertificate API in support of managed renewal
* `service/health`: Updates service documentation

Release v1.6.14 (2017-01-19)
===

Service Client Updates
---
* `service/ec2`: Updates service API, documentation, and paginators
  * Amazon EC2 Spot instances now support dedicated tenancy, providing the ability to run Spot instances single-tenant manner on physically isolated hardware within a VPC to satisfy security, privacy, or other compliance requirements. Dedicated Spot instances can be requested using RequestSpotInstances and RequestSpotFleet.

Release v1.6.13 (2017-01-18)
===

Service Client Updates
---
* `service/rds`: Updates service API, documentation, and paginators

Release v1.6.12 (2017-01-17)
===

Service Client Updates
---
* `service/dynamodb`: Updates service API, documentation, and paginators
  * Tagging Support for Amazon DynamoDB Tables and Indexes
* `aws/endpoints`: Updated Regions and Endpoints metadata.
* `service/glacier`: Updates service API, paginators, and examples
  * Doc-only Update for Glacier: Added code snippets
* `service/polly`: Updates service documentation and examples
  * Doc-only update for Amazon Polly -- added snippets
* `service/rekognition`: Updates service documentation and paginators
  * Added code samples to Rekognition reference topics.
* `service/route53`: Updates service API and paginators
  * Add ca-central-1 and eu-west-2 enum values to CloudWatchRegion enum

Release v1.6.11 (2017-01-16)
===

Service Client Updates
---
* `service/configservice`: Updates service API, documentation, and paginators
* `service/costandusagereportservice`: Adds new service
  * The AWS Cost and Usage Report Service API allows you to enable and disable the Cost & Usage report, as well as modify the report name, the data granularity, and the delivery preferences.
* `service/dynamodb`: Updates service API, documentation, and examples
  * Snippets for the DynamoDB API.
* `service/elasticache`: Updates service API, documentation, and examples
  * Adds new code examples.
* `aws/endpoints`: Updated Regions and Endpoints metadata.

Release v1.6.10 (2017-01-04)
===

Service Client Updates
---
* `service/configservice`: Updates service API and documentation
  * AWSConfig is planning to add support for OversizedConfigurationItemChangeNotification message type in putConfigRule. After this release customers can use/write rules based on OversizedConfigurationItemChangeNotification mesage type.
* `service/efs`: Updates service API, documentation, and examples
  * Doc-only Update for EFS: Added code snippets
* `service/iam`: Updates service documentation and examples
* `service/lambda`: Updates service documentation and examples
  * Doc only updates for Lambda: Added code snippets
* `service/marketplacecommerceanalytics`: Updates service API and documentation
  * Added support for data set disbursed_amount_by_instance_hours, with historical data available starting 2012-09-04. New data is published to this data set every 30 days.
* `service/rds`: Updates service documentation
  * Updated documentation for CopyDBSnapshot.
* `service/rekognition`: Updates service documentation and examples
  * Doc-only Update for Rekognition: Added code snippets
* `service/snowball`: Updates service examples
* `service/dynamodbstreams`: Updates service API and examples
  * Doc-only Update for DynamoDB Streams:  Added code snippets

SDK Feature
---
* `private/model/api`: Increasing the readability of code generated files. (#1024)
Release v1.6.9 (2016-12-30)
===

Service Client Updates
---
* `service/codedeploy`: Updates service API and documentation
  * CodeDeploy will support Iam Session Arns in addition to Iam User Arns for on premise host authentication.
* `service/ecs`: Updates service API and documentation
  * Amazon EC2 Container Service (ECS) now supports the ability to customize the placement of tasks on container instances.
* `aws/endpoints`: Updated Regions and Endpoints metadata.

Release v1.6.8 (2016-12-22)
===

Service Client Updates
---
* `service/apigateway`: Updates service API and documentation
  * Amazon API Gateway is adding support for generating SDKs in more languages. This update introduces two new operations used to dynamically discover these SDK types and what configuration each type accepts.
* `service/directoryservice`: Updates service documentation
  * Added code snippets for the DS SDKs
* `service/elasticbeanstalk`: Updates service API and documentation
* `service/iam`: Updates service API and documentation
  * Adds service-specific credentials to IAM service to make it easier to onboard CodeCommit customers.  These are username/password credentials that work with a single service.
* `service/kms`: Updates service API, documentation, and examples
  * Update docs and add SDK examples

Release v1.6.7 (2016-12-22)
===

Service Client Updates
---
* `service/ecr`: Updates service API and documentation
* `aws/endpoints`: Updated Regions and Endpoints metadata.
* `service/rds`: Updates service API and documentation
  * Cross Region Encrypted Snapshot Copying (CopyDBSnapshot)

Release v1.6.6 (2016-12-20)
===

Service Client Updates
---
* `aws/endpoints`: Updated Regions and Endpoints metadata.
* `service/firehose`: Updates service API, documentation, and examples
  * Processing feature enables users to process and modify records before Amazon Firehose delivers them to destinations.
* `service/route53`: Updates service API and documentation
  * Enum updates for eu-west-2 and ca-central-1
* `service/storagegateway`: Updates service API, documentation, and examples
  * File gateway is a new mode in the AWS Storage Gateway that support a file interface into S3, alongside the current block-based volume and VTL storage. File gateway combines a service and virtual software appliance, enabling you to store and retrieve objects in Amazon S3 using industry standard file protocols such as NFS. The software appliance, or gateway, is deployed into your on-premises environment as a virtual machine (VM) running on VMware ESXi. The gateway provides access to objects in S3 as files on a Network File System (NFS) mount point.

Release v1.6.5 (2016-12-19)
===

Service Client Updates
---
* `service/cloudformation`: Updates service documentation
  * Minor doc update for CloudFormation.
* `service/cloudtrail`: Updates service paginators
* `service/cognitoidentity`: Updates service API and documentation
  * We are adding Groups to Cognito user pools. Developers can perform CRUD operations on groups, add and remove users from groups, list users in groups, etc. We are adding fine-grained role-based access control for Cognito identity pools. Developers can configure an identity pool to get the IAM role from an authenticated user's token, or they can configure rules that will map a user to a different role
* `service/applicationdiscoveryservice`: Updates service API and documentation
  * Adds new APIs to group discovered servers into Applications with get summary and neighbors. Includes additional filters for ListConfigurations and DescribeAgents API.
* `service/inspector`: Updates service API, documentation, and examples
  * Doc-only Update for Inspector: Adding SDK code snippets for Inspector
* `service/sqs`: Updates service documentation

SDK Bug Fixes
---
* `aws/request`: Add PriorRequestNotComplete to throttle retry codes (#1011)
  * Fixes: Not retrying when PriorRequestNotComplete #1009

SDK Feature
---
* `private/model/api`: Adds crosslinking to service documentation (#1010)

Release v1.6.4 (2016-12-15)
===

Service Client Updates
---
* `service/cognitoidentityprovider`: Updates service API and documentation
* `aws/endpoints`: Updated Regions and Endpoints metadata.
* `service/ssm`: Updates service API and documentation
  * This will provide customers with access to the Patch Baseline and Patch Compliance APIs.

SDK Bug Fixes
---
* `service/route53`: Fix URL path cleaning for Route53 API requests (#1006)
  * Fixes: SerializationError when using Route53 ChangeResourceRecordSets #1005
* `aws/request`: Add PriorRequestNotComplete to throttle retry codes (#1002)
  * Fixes: Not retrying when PriorRequestNotComplete #1001

Release v1.6.3 (2016-12-14)
===

Service Client Updates
---
* `service/batch`: Adds new service
  * AWS Batch is a batch computing service that lets customers define queues and compute environments and then submit work as batch jobs.
* `service/databasemigrationservice`: Updates service API and documentation
  * Adds support for SSL enabled Oracle endpoints and task modification.
* `service/elasticbeanstalk`: Updates service documentation
* `aws/endpoints`: Updated Regions and Endpoints metadata.
* `service/cloudwatchlogs`: Updates service API and documentation
  * Add support for associating LogGroups with AWSTagris tags
* `service/marketplacecommerceanalytics`: Updates service API and documentation
  * Add new enum to DataSetType: sales_compensation_billed_revenue
* `service/rds`: Updates service documentation
  * Doc-only Update for RDS: New versions available in CreateDBInstance
* `service/sts`: Updates service documentation
  * Adding Code Snippet Examples for SDKs for STS

SDK Bug Fixes
---
* `aws/request`: Fix retrying timeout requests (#981)
  * Fixes: Requests Retrying is broken if the error was caused due to a client timeout #947
* `aws/request`: Fix for Go 1.8 request incorrectly sent with body (#991)
  * Fixes: service/route53: ListHostedZones hangs and then fails with go1.8 #984
* private/protocol/rest: Use RawPath instead of Opaque (#993)
  * Fixes: HTTP2 request failing with REST protocol services, e.g AWS X-Ray
* private/model/api: Generate REST-JSON JSONVersion correctly (#998)
  * Fixes: REST-JSON protocol service code missing JSONVersion metadata.

Release v1.6.2 (2016-12-08)
===

Service Client Updates
---
* `service/cloudfront`: Add lambda function associations to cache behaviors
* `service/codepipeline`: This is a doc-only update request to incorporate some recent minor revisions to the doc content.
* `service/rds`: Updates service API and documentation
* `service/wafregional`: With this new feature, customers can use AWS WAF directly on Application Load Balancers in a VPC within available regions to protect their websites and web services from malicious attacks such as SQL injection, Cross Site Scripting, bad bots, etc.

Release v1.6.1 (2016-12-07)
===

Service Client Updates
---
* `service/config`: Updates service API
* `service/s3`: Updates service API
* `service/sqs`: Updates service API and documentation

Release v1.6.0 (2016-12-06)
===

Service Client Updates
---
* `service/config`: Updates service API and documentation
* `service/ec2`: Updates service API
* `service/sts`: Updates service API, documentation, and examples

SDK Bug Fixes
---
* private/protocol/xml/xmlutil: Fix SDK XML unmarshaler #975
  * Fixes GetBucketACL Grantee required type always nil. #916

SDK Feature
---
* aws/endpoints: Add endpoint metadata to SDK #961
  * Adds Region and Endpoint metadata to the SDK. This allows you to enumerate regions and endpoint metadata based on a defined model embedded in the SDK.

Release v1.5.13 (2016-12-01)
===

Service Client Updates
---
* `service/apigateway`: Updates service API and documentation
* `service/appstream`: Adds new service
* `service/codebuild`: Adds new service
* `service/directconnect`: Updates service API and documentation
* `service/ec2`: Adds new service
* `service/elasticbeanstalk`: Updates service API and documentation
* `service/health`: Adds new service
* `service/lambda`: Updates service API and documentation
* `service/opsworkscm`: Adds new service
* `service/pinpoint`: Adds new service
* `service/shield`: Adds new service
* `service/ssm`: Updates service API and documentation
* `service/states`: Adds new service
* `service/xray`: Adds new service

Release v1.5.12 (2016-11-30)
===

Service Client Updates
---
* `service/lightsail`: Adds new service
* `service/polly`: Adds new service
* `service/rekognition`: Adds new service
* `service/snowball`: Updates service API and documentation

Release v1.5.11 (2016-11-29)
===

Service Client Updates
---
`service/s3`: Updates service API and documentation

Release v1.5.10 (2016-11-22)
===

Service Client Updates
---
* `service/cloudformation`: Updates service API and documentation
* `service/glacier`: Updates service API, documentation, and examples
* `service/route53`: Updates service API and documentation
* `service/s3`: Updates service API and documentation

SDK Bug Fixes
---
* `private/protocol/xml/xmlutil`: Fixes xml marshaler to unmarshal properly
into tagged fields 
[#916](https://github.com/aws/aws-sdk-go/issues/916)

Release v1.5.9 (2016-11-22)
===

Service Client Updates
---
* `service/cloudtrail`: Updates service API and documentation
* `service/ecs`: Updates service API and documentation

Release v1.5.8 (2016-11-18)
===

Service Client Updates
---
* `service/application-autoscaling`: Updates service API and documentation
* `service/elasticmapreduce`: Updates service API and documentation
* `service/elastictranscoder`: Updates service API, documentation, and examples
* `service/gamelift`: Updates service API and documentation
* `service/lambda`: Updates service API and documentation

Release v1.5.7 (2016-11-18)
===

Service Client Updates
---
* `service/apigateway`: Updates service API and documentation
* `service/meteringmarketplace`: Updates service API and documentation
* `service/monitoring`: Updates service API and documentation
* `service/sqs`: Updates service API, documentation, and examples

Release v1.5.6 (2016-11-16)
===

Service Client Updates
---
`service/route53`: Updates service API and documentation
`service/servicecatalog`: Updates service API and documentation

Release v1.5.5 (2016-11-15)
===

Service Client Updates
---
* `service/ds`: Updates service API and documentation
* `service/elasticache`: Updates service API and documentation
* `service/kinesis`: Updates service API and documentation

Release v1.5.4 (2016-11-15)
===

Service Client Updates
---
* `service/cognito-idp`: Updates service API and documentation

Release v1.5.3 (2016-11-11)
===

Service Client Updates
---
* `service/cloudformation`: Updates service documentation and examples
* `service/logs`: Updates service API and documentation

Release v1.5.2 (2016-11-03)
===

Service Client Updates
---
* `service/directconnect`: Updates service API and documentation

Release v1.5.1 (2016-11-02)
===

Service Client Updates
---
* `service/email`: Updates service API and documentation

Release v1.5.0 (2016-11-01)
===

Service Client Updates
---
* `service/cloudformation`: Updates service API and documentation
* `service/ecr`: Updates service paginators

SDK Feature Updates
---
* `private/model/api`: Add generated setters for API parameters (#918)
  * Adds setters to the SDK's API parameter types, and are a convenience method that reduce the need to use `aws.String` and like utility. 

Release v1.4.22 (2016-10-25)
===

Service Client Updates
---
* `service/elasticloadbalancingv2`: Updates service documentation.
* `service/autoscaling`: Updates service documentation.

Release v1.4.21 (2016-10-24)
===

Service Client Updates
---
* `service/sms`: AWS Server Migration Service (SMS) is an agentless service which makes it easier and faster for you to migrate thousands of on-premises workloads to AWS. AWS SMS allows you to automate, schedule, and track incremental replications of live server volumes, making it easier for you to coordinate large-scale server migrations.
* `service/ecs`: Updates documentation.

SDK Feature Updates
---
* `private/models/api`: Improve code generation of documentation.

Release v1.4.20 (2016-10-20)
===

Service Client Updates
---
* `service/budgets`: Adds new service, AWS Budgets.
* `service/waf`: Updates service documentation.

Release v1.4.19 (2016-10-18)
===

Service Client Updates
---
* `service/cloudfront`: Updates service API and documentation.
  * Ability to use Amazon CloudFront to deliver your content both via IPv6 and IPv4 using HTTP/HTTPS.
* `service/configservice`: Update service API and documentation.
* `service/iot`: Updates service API and documentation.
* `service/kinesisanalytics`: Updates service API and documentation.
  * Whenever Amazon Kinesis Analytics is not able to detect schema for the given streaming source on DiscoverInputSchema API, we would return the raw records that was sampled to detect the schema.
* `service/rds`: Updates service API and documentation.
  * Amazon Aurora integrates with other AWS services to allow you to extend your Aurora DB cluster to utilize other capabilities in the AWS cloud. Permission to access other AWS services is granted by creating an IAM role with the necessary permissions, and then associating the role with your DB cluster.

SDK Feature Updates
---
* `service/dynamodb/dynamodbattribute`: Add UnmarshalListOfMaps #897
  * Adds support for unmarshaling a list of maps. This is useful for unmarshaling the DynamoDB AttributeValue list of maps returned by APIs like Query and Scan.

Release v1.4.18 (2016-10-17)
===

Service Model Updates
---
* `service/route53`: Updates service API and documentation.

Release v1.4.17
===

Service Model Updates
---
* `service/acm`: Update service API, and documentation.
  * This change allows users to import third-party SSL/TLS certificates into ACM.
* `service/elasticbeanstalk`: Update service API, documentation, and pagination.
  * Elastic Beanstalk DescribeApplicationVersions API is being updated to support pagination.
* `service/gamelift`: Update service API, and documentation.
  * New APIs to protect game developer resource (builds, alias, fleets, instances, game sessions and player sessions) against abuse.

SDK Features
---
* `service/s3`: Add support for accelerate with dualstack [#887](https://github.com/aws/aws-sdk-go/issues/887)

Release v1.4.16 (2016-10-13)
===

Service Model Updates
---
* `service/ecr`: Update Amazon EC2 Container Registry service model
  * DescribeImages is a new api used to expose image metadata which today includes image size and image creation timestamp.
* `service/elasticache`: Update Amazon ElastiCache service model
  * Elasticache is launching a new major engine release of Redis, 3.2 (providing stability updates and new command sets over 2.8), as well as ElasticSupport for enabling Redis Cluster in 3.2, which provides support for multiple node groups to horizontally scale data, as well as superior engine failover capabilities 

SDK Bug Fixes
---
* `aws/session`: Skip shared config on read errors [#883](https://github.com/aws/aws-sdk-go/issues/883)
* `aws/signer/v4`: Add support for URL.EscapedPath to signer [#885](https://github.com/aws/aws-sdk-go/issues/885)

SDK Features
---
* `private/model/api`: Add docs for errors to API operations [#881](https://github.com/aws/aws-sdk-go/issues/881)
* `private/model/api`: Improve field and waiter doc strings [#879](https://github.com/aws/aws-sdk-go/issues/879)
* `service/dynamodb/dynamodbattribute`: Allow multiple struct tag elements [#886](https://github.com/aws/aws-sdk-go/issues/886)
* Add build tags to internal SDK tools [#880](https://github.com/aws/aws-sdk-go/issues/880)

Release v1.4.15 (2016-10-06)
===

Service Model Updates
---
* `service/cognitoidentityprovider`: Update Amazon Cognito Identity Provider service model
* `service/devicefarm`: Update AWS Device Farm documentation
* `service/opsworks`: Update AWS OpsWorks service model
* `service/s3`: Update Amazon Simple Storage Service model
* `service/waf`: Update AWS WAF service model

SDK Bug Fixes
---
* `aws/request`: Fix HTTP Request Body race condition [#874](https://github.com/aws/aws-sdk-go/issues/874)

SDK Feature Updates
---
* `aws/ec2metadata`: Add support for EC2 User Data [#872](https://github.com/aws/aws-sdk-go/issues/872)
* `aws/signer/v4`: Remove logic determining if request needs to be resigned [#876](https://github.com/aws/aws-sdk-go/issues/876)

Release v1.4.14 (2016-09-29)
===
* `service/ec2`:  api, documentation, and paginators updates.
* `service/s3`:  api and documentation updates.

Release v1.4.13 (2016-09-27)
===
* `service/codepipeline`:  documentation updates.
* `service/cloudformation`:  api and documentation updates.
* `service/kms`:  documentation updates.
* `service/elasticfilesystem`:  documentation updates.
* `service/snowball`:  documentation updates.
