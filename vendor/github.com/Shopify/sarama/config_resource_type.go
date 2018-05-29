package sarama

type ConfigResourceType int8

// Taken from :
// https://cwiki.apache.org/confluence/display/KAFKA/KIP-133%3A+Describe+and+Alter+Configs+Admin+APIs#KIP-133:DescribeandAlterConfigsAdminAPIs-WireFormattypes

const (
	UnknownResource ConfigResourceType = 0
	AnyResource     ConfigResourceType = 1
	TopicResource   ConfigResourceType = 2
	GroupResource   ConfigResourceType = 3
	ClusterResource ConfigResourceType = 4
	BrokerResource  ConfigResourceType = 5
)
