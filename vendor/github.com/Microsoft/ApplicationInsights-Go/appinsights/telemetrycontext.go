package appinsights

import (
	"os"
	"runtime"
	"strconv"
)

type TelemetryContext interface {
	InstrumentationKey() string
	loadDeviceContext()

	Component() ComponentContext
	Device() DeviceContext
	Cloud() CloudContext
	Session() SessionContext
	User() UserContext
	Operation() OperationContext
	Location() LocationContext
}

type telemetryContext struct {
	iKey string
	tags map[string]string
}

type ComponentContext interface {
	GetVersion() string
	SetVersion(string)
}

type DeviceContext interface {
	GetType() string
	SetType(string)
	GetId() string
	SetId(string)
	GetOperatingSystem() string
	SetOperatingSystem(string)
	GetOemName() string
	SetOemName(string)
	GetModel() string
	SetModel(string)
	GetNetworkType() string
	SetNetworkType(string)
	GetScreenResolution() string
	SetScreenResolution(string)
	GetLanguage() string
	SetLanguage(string)
}

type CloudContext interface {
	GetRoleName() string
	SetRoleName(string)
	GetRoleInstance() string
	SetRoleInstance(string)
}

type SessionContext interface {
	GetId() string
	SetId(string)
	GetIsFirst() bool
	SetIsFirst(bool)
}

type UserContext interface {
	GetId() string
	SetId(string)
	GetAccountId() string
	SetAccountId(string)
	GetUserAgent() string
	SetUserAgent(string)
	GetAuthenticatedUserId() string
	SetAuthenticatedUserId(string)
}

type OperationContext interface {
	GetId() string
	SetId(string)
	GetParentId() string
	SetParentId(string)
	GetCorrelationVector() string
	SetCorrelationVector(string)
	GetName() string
	SetName(string)
	GetSyntheticSource() string
	SetSyntheticSource(string)
}

type LocationContext interface {
	GetIp() string
	SetIp(string)
}

func NewItemTelemetryContext() TelemetryContext {
	context := &telemetryContext{
		tags: make(map[string]string),
	}
	return context
}

func NewClientTelemetryContext() TelemetryContext {
	context := &telemetryContext{
		tags: make(map[string]string),
	}
	context.loadDeviceContext()
	context.loadInternalContext()
	return context
}

func (context *telemetryContext) InstrumentationKey() string {
	return context.iKey
}

func (context *telemetryContext) loadDeviceContext() {
	hostname, err := os.Hostname()
	if err == nil {
		context.tags[DeviceId] = hostname
		context.tags[DeviceMachineName] = hostname
		context.tags[DeviceRoleInstance] = hostname
	}
	context.tags[DeviceOS] = runtime.GOOS
}

func (context *telemetryContext) loadInternalContext() {
	context.tags[InternalSdkVersion] = sdkName + ":" + Version
}

func (context *telemetryContext) Component() ComponentContext {
	return &componentContext{context: context}
}

func (context *telemetryContext) Device() DeviceContext {
	return &deviceContext{context: context}
}

func (context *telemetryContext) Cloud() CloudContext {
	return &cloudContext{context: context}
}

func (context *telemetryContext) Session() SessionContext {
	return &sessionContext{context: context}
}

func (context *telemetryContext) User() UserContext {
	return &userContext{context: context}
}

func (context *telemetryContext) Operation() OperationContext {
	return &operationContext{context: context}
}

func (context *telemetryContext) Location() LocationContext {
	return &locationContext{context: context}
}

func (context *telemetryContext) getTagString(key ContextTagKeys) string {
	if val, ok := context.tags[string(key)]; ok {
		return val
	}

	return ""
}

func (context *telemetryContext) setTagString(key ContextTagKeys, value string) {
	if value != "" {
		context.tags[string(key)] = value
	} else {
		delete(context.tags, string(key))
	}
}

func (context *telemetryContext) getTagBool(key ContextTagKeys) bool {
	if val, ok := context.tags[string(key)]; ok {
		if b, err := strconv.ParseBool(val); err != nil {
			return b
		}
	}

	return false
}

func (context *telemetryContext) setTagBool(key ContextTagKeys, value bool) {
	if value {
		context.tags[string(key)] = "true"
	} else {
		delete(context.tags, string(key))
	}
}

type componentContext struct {
	context *telemetryContext
}

type deviceContext struct {
	context *telemetryContext
}

type cloudContext struct {
	context *telemetryContext
}

type sessionContext struct {
	context *telemetryContext
}

type userContext struct {
	context *telemetryContext
}

type operationContext struct {
	context *telemetryContext
}

type locationContext struct {
	context *telemetryContext
}

func (context *componentContext) GetVersion() string {
	return context.context.getTagString(ApplicationVersion)
}

func (context *componentContext) SetVersion(value string) {
	context.context.setTagString(ApplicationVersion, value)
}

func (context *deviceContext) GetType() string {
	return context.context.getTagString(DeviceType)
}

func (context *deviceContext) SetType(value string) {
	context.context.setTagString(DeviceType, value)
}

func (context *deviceContext) GetId() string {
	return context.context.getTagString(DeviceId)
}

func (context *deviceContext) SetId(value string) {
	context.context.setTagString(DeviceId, value)
}

func (context *deviceContext) GetOperatingSystem() string {
	return context.context.getTagString(DeviceOSVersion)
}

func (context *deviceContext) SetOperatingSystem(value string) {
	context.context.setTagString(DeviceOSVersion, value)
}

func (context *deviceContext) GetOemName() string {
	return context.context.getTagString(DeviceOEMName)
}

func (context *deviceContext) SetOemName(value string) {
	context.context.setTagString(DeviceOEMName, value)
}

func (context *deviceContext) GetModel() string {
	return context.context.getTagString(DeviceModel)
}

func (context *deviceContext) SetModel(value string) {
	context.context.setTagString(DeviceModel, value)
}

func (context *deviceContext) GetNetworkType() string {
	return context.context.getTagString(DeviceNetwork)
}

func (context *deviceContext) SetNetworkType(value string) {
	context.context.setTagString(DeviceNetwork, value)
}

func (context *deviceContext) GetScreenResolution() string {
	return context.context.getTagString(DeviceScreenResolution)
}

func (context *deviceContext) SetScreenResolution(value string) {
	context.context.setTagString(DeviceScreenResolution, value)
}

func (context *deviceContext) GetLanguage() string {
	return context.context.getTagString(DeviceLanguage)
}

func (context *deviceContext) SetLanguage(value string) {
	context.context.setTagString(DeviceLanguage, value)
}

func (context *cloudContext) GetRoleName() string {
	return context.context.getTagString(CloudRole)
}

func (context *cloudContext) SetRoleName(value string) {
	context.context.setTagString(CloudRole, value)
}

func (context *cloudContext) GetRoleInstance() string {
	return context.context.getTagString(CloudRoleInstance)
}

func (context *cloudContext) SetRoleInstance(value string) {
	context.context.setTagString(CloudRoleInstance, value)
}

func (context *sessionContext) GetId() string {
	return context.context.getTagString(SessionId)
}

func (context *sessionContext) SetId(value string) {
	context.context.setTagString(SessionId, value)
}

func (context *sessionContext) GetIsFirst() bool {
	return context.context.getTagBool(SessionIsFirst)
}

func (context *sessionContext) SetIsFirst(value bool) {
	context.context.setTagBool(SessionIsFirst, value)
}

func (context *userContext) GetId() string {
	return context.context.getTagString(UserId)
}

func (context *userContext) SetId(value string) {
	context.context.setTagString(UserId, value)
}

func (context *userContext) GetAccountId() string {
	return context.context.getTagString(UserAccountId)
}

func (context *userContext) SetAccountId(value string) {
	context.context.setTagString(UserAccountId, value)
}

func (context *userContext) GetUserAgent() string {
	return context.context.getTagString(UserAgent)
}

func (context *userContext) SetUserAgent(value string) {
	context.context.setTagString(UserAgent, value)
}

func (context *userContext) GetAuthenticatedUserId() string {
	return context.context.getTagString(UserAuthUserId)
}

func (context *userContext) SetAuthenticatedUserId(value string) {
	context.context.setTagString(UserAuthUserId, value)
}

func (context *operationContext) GetId() string {
	return context.context.getTagString(OperationId)
}

func (context *operationContext) SetId(value string) {
	context.context.setTagString(OperationId, value)
}

func (context *operationContext) GetParentId() string {
	return context.context.getTagString(OperationParentId)
}

func (context *operationContext) SetParentId(value string) {
	context.context.setTagString(OperationParentId, value)
}

func (context *operationContext) GetCorrelationVector() string {
	return context.context.getTagString(OperationCorrelationVector)
}

func (context *operationContext) SetCorrelationVector(value string) {
	context.context.setTagString(OperationCorrelationVector, value)
}

func (context *operationContext) GetName() string {
	return context.context.getTagString(OperationName)
}

func (context *operationContext) SetName(value string) {
	context.context.setTagString(OperationName, value)
}

func (context *operationContext) GetSyntheticSource() string {
	return context.context.getTagString(OperationSyntheticSource)
}

func (context *operationContext) SetSyntheticSource(value string) {
	context.context.setTagString(OperationSyntheticSource, value)
}

func (context *locationContext) GetIp() string {
	return context.context.getTagString(LocationIp)
}

func (context *locationContext) SetIp(value string) {
	context.context.setTagString(LocationIp, value)
}
