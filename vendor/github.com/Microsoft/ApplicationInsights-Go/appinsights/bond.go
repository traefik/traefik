package appinsights

type Domain interface {
}

type domain struct {
	Ver        int               `json:"ver"`
	Properties map[string]string `json:"properties"`
}

type data struct {
	BaseType string `json:"baseType"`
	BaseData Domain `json:"baseData"`
}

type envelope struct {
	Name string            `json:"name"`
	Time string            `json:"time"`
	IKey string            `json:"iKey"`
	Tags map[string]string `json:"tags"`
	Data *data             `json:"data"`
}

type DataPointType int

const (
	Measurement DataPointType = iota
	Aggregation
)

type DataPoint struct {
	Name   string        `json:"name"`
	Kind   DataPointType `json:"kind"`
	Value  float32       `json:"value"`
	Count  int           `json:"count"`
	min    float32       `json:"min"`
	max    float32       `json:"max"`
	stdDev float32       `json:"stdDev"`
}

type metricData struct {
	domain
	Metrics []*DataPoint `json:"metrics"`
}

type eventData struct {
	domain
	Name         string             `json:"name"`
	Measurements map[string]float32 `json:"measurements"`
}

type SeverityLevel int

const (
	Verbose SeverityLevel = iota
	Information
	Warning
	Error
	Critical
)

type messageData struct {
	domain
	Message       string        `json:"message"`
	SeverityLevel SeverityLevel `json:"severityLevel"`
}

type requestData struct {
	domain
	Id           string             `json:"id"`
	Name         string             `json:"name"`
	StartTime    string             `json:"startTime"` // yyyy-mm-ddThh:mm:ss.fffffff-hh:mm
	Duration     string             `json:"duration"`  // d:hh:mm:ss.fffffff
	ResponseCode string             `json:"responseCode"`
	Success      bool               `json:"success"`
	HttpMethod   string             `json:"httpMethod"`
	Url          string             `json:"url"`
	Measurements map[string]float32 `json:"measurements"`
}

type ContextTagKeys string

const (
	ApplicationVersion         ContextTagKeys = "ai.application.ver"
	ApplicationBuild                          = "ai.application.build"
	CloudRole                                 = "ai.cloud.role"
	CloudRoleInstance                         = "ai.cloud.roleInstance"
	DeviceId                                  = "ai.device.id"
	DeviceIp                                  = "ai.device.ip"
	DeviceLanguage                            = "ai.device.language"
	DeviceLocale                              = "ai.device.locale"
	DeviceModel                               = "ai.device.model"
	DeviceNetwork                             = "ai.device.network"
	DeviceOEMName                             = "ai.device.oemName"
	DeviceOS                                  = "ai.device.os"
	DeviceOSVersion                           = "ai.device.osVersion"
	DeviceRoleInstance                        = "ai.device.roleInstance"
	DeviceRoleName                            = "ai.device.roleName"
	DeviceScreenResolution                    = "ai.device.screenResolution"
	DeviceType                                = "ai.device.type"
	DeviceMachineName                         = "ai.device.machineName"
	LocationIp                                = "ai.location.ip"
	OperationCorrelationVector                = "ai.operation.correlationVector"
	OperationId                               = "ai.operation.id"
	OperationName                             = "ai.operation.name"
	OperationParentId                         = "ai.operation.parentId"
	OperationRootId                           = "ai.operation.rootId"
	OperationSyntheticSource                  = "ai.operation.syntheticSource"
	OperationIsSynthetic                      = "ai.operation.isSynthetic"
	SessionId                                 = "ai.session.id"
	SessionIsFirst                            = "ai.session.isFirst"
	SessionIsNew                              = "ai.session.isNew"
	UserAccountAcquisitionDate                = "ai.user.accountAcquisitionDate"
	UserAccountId                             = "ai.user.accountId"
	UserAgent                                 = "ai.user.userAgent"
	UserAuthUserId                            = "ai.user.authUserId"
	UserId                                    = "ai.user.id"
	UserStoreRegion                           = "ai.user.storeRegion"
	SampleRate                                = "ai.sample.sampleRate"
	InternalSdkVersion                        = "ai.internal.sdkVersion"
	InternalAgentVersion                      = "ai.internal.agentVersion"
)
