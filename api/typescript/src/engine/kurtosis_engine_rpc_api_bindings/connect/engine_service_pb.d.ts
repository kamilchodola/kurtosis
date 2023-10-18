// @generated by protoc-gen-es v1.3.0 with parameter "target=js+dts"
// @generated from file engine_service.proto (package engine_api, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import type { BinaryReadOptions, FieldList, JsonReadOptions, JsonValue, PartialMessage, PlainMessage, Timestamp } from "@bufbuild/protobuf";
import { Message, proto3 } from "@bufbuild/protobuf";

/**
 * @generated from enum engine_api.EnclaveMode
 */
export declare enum EnclaveMode {
  /**
   * @generated from enum value: TEST = 0;
   */
  TEST = 0,

  /**
   * @generated from enum value: PRODUCTION = 1;
   */
  PRODUCTION = 1,
}

/**
 * ==============================================================================================
 *                                            Get Enclaves
 * ==============================================================================================
 * Status of the containers in the enclave
 * NOTE: We have to prefix the enum values with the enum name due to the way Protobuf enum valuee uniqueness works
 *
 * @generated from enum engine_api.EnclaveContainersStatus
 */
export declare enum EnclaveContainersStatus {
  /**
   * The enclave has been created, but there are no containers inside it
   *
   * @generated from enum value: EnclaveContainersStatus_EMPTY = 0;
   */
  EnclaveContainersStatus_EMPTY = 0,

  /**
   * One or more containers are running in the enclave (which may or may not include the API container, depending on if the user was manually stopping/removing containers)
   *
   * @generated from enum value: EnclaveContainersStatus_RUNNING = 1;
   */
  EnclaveContainersStatus_RUNNING = 1,

  /**
   * There are >= 1 container in the enclave, but they're all stopped
   *
   * @generated from enum value: EnclaveContainersStatus_STOPPED = 2;
   */
  EnclaveContainersStatus_STOPPED = 2,
}

/**
 * NOTE: We have to prefix the enum values with the enum name due to the way Protobuf enum value uniqueness works
 *
 * @generated from enum engine_api.EnclaveAPIContainerStatus
 */
export declare enum EnclaveAPIContainerStatus {
  /**
   * No API container exists in the enclave
   * This is the only valid value when the enclave containers status is "EMPTY"
   *
   * @generated from enum value: EnclaveAPIContainerStatus_NONEXISTENT = 0;
   */
  EnclaveAPIContainerStatus_NONEXISTENT = 0,

  /**
   * An API container exists and is running
   * NOTE: this does NOT say that the server inside the API container is available, because checking if it's available requires making a call to the API container
   *  If we have a lot of API containers, we'd be making tons of calls
   *
   * @generated from enum value: EnclaveAPIContainerStatus_RUNNING = 1;
   */
  EnclaveAPIContainerStatus_RUNNING = 1,

  /**
   * An API container exists, but isn't running
   *
   * @generated from enum value: EnclaveAPIContainerStatus_STOPPED = 2;
   */
  EnclaveAPIContainerStatus_STOPPED = 2,
}

/**
 * The filter operator which can be text or regex type
 * NOTE: We have to prefix the enum values with the enum name due to the way Protobuf enum value uniqueness works
 *
 * @generated from enum engine_api.LogLineOperator
 */
export declare enum LogLineOperator {
  /**
   * @generated from enum value: LogLineOperator_DOES_CONTAIN_TEXT = 0;
   */
  LogLineOperator_DOES_CONTAIN_TEXT = 0,

  /**
   * @generated from enum value: LogLineOperator_DOES_NOT_CONTAIN_TEXT = 1;
   */
  LogLineOperator_DOES_NOT_CONTAIN_TEXT = 1,

  /**
   * @generated from enum value: LogLineOperator_DOES_CONTAIN_MATCH_REGEX = 2;
   */
  LogLineOperator_DOES_CONTAIN_MATCH_REGEX = 2,

  /**
   * @generated from enum value: LogLineOperator_DOES_NOT_CONTAIN_MATCH_REGEX = 3;
   */
  LogLineOperator_DOES_NOT_CONTAIN_MATCH_REGEX = 3,
}

/**
 * ==============================================================================================
 *                                        Get Engine Info
 * ==============================================================================================
 *
 * @generated from message engine_api.GetEngineInfoResponse
 */
export declare class GetEngineInfoResponse extends Message<GetEngineInfoResponse> {
  /**
   * Version of the engine server
   *
   * @generated from field: string engine_version = 1;
   */
  engineVersion: string;

  constructor(data?: PartialMessage<GetEngineInfoResponse>);

  static readonly runtime: typeof proto3;
  static readonly typeName = "engine_api.GetEngineInfoResponse";
  static readonly fields: FieldList;

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetEngineInfoResponse;

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetEngineInfoResponse;

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetEngineInfoResponse;

  static equals(a: GetEngineInfoResponse | PlainMessage<GetEngineInfoResponse> | undefined, b: GetEngineInfoResponse | PlainMessage<GetEngineInfoResponse> | undefined): boolean;
}

/**
 * ==============================================================================================
 *                                        Create Enclave
 * ==============================================================================================
 *
 * @generated from message engine_api.CreateEnclaveArgs
 */
export declare class CreateEnclaveArgs extends Message<CreateEnclaveArgs> {
  /**
   * The name of the new Kurtosis Enclave
   *
   * @generated from field: optional string enclave_name = 1;
   */
  enclaveName?: string;

  /**
   * The image tag of the API container that should be used inside the enclave
   * If blank, will use the default version that the engine server uses
   *
   * @generated from field: optional string api_container_version_tag = 2;
   */
  apiContainerVersionTag?: string;

  /**
   * The API container log level
   *
   * @generated from field: optional string api_container_log_level = 3;
   */
  apiContainerLogLevel?: string;

  /**
   * @generated from field: optional engine_api.EnclaveMode mode = 4;
   */
  mode?: EnclaveMode;

  constructor(data?: PartialMessage<CreateEnclaveArgs>);

  static readonly runtime: typeof proto3;
  static readonly typeName = "engine_api.CreateEnclaveArgs";
  static readonly fields: FieldList;

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): CreateEnclaveArgs;

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): CreateEnclaveArgs;

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): CreateEnclaveArgs;

  static equals(a: CreateEnclaveArgs | PlainMessage<CreateEnclaveArgs> | undefined, b: CreateEnclaveArgs | PlainMessage<CreateEnclaveArgs> | undefined): boolean;
}

/**
 * @generated from message engine_api.CreateEnclaveResponse
 */
export declare class CreateEnclaveResponse extends Message<CreateEnclaveResponse> {
  /**
   * All the enclave information inside this object
   *
   * @generated from field: engine_api.EnclaveInfo enclave_info = 1;
   */
  enclaveInfo?: EnclaveInfo;

  constructor(data?: PartialMessage<CreateEnclaveResponse>);

  static readonly runtime: typeof proto3;
  static readonly typeName = "engine_api.CreateEnclaveResponse";
  static readonly fields: FieldList;

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): CreateEnclaveResponse;

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): CreateEnclaveResponse;

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): CreateEnclaveResponse;

  static equals(a: CreateEnclaveResponse | PlainMessage<CreateEnclaveResponse> | undefined, b: CreateEnclaveResponse | PlainMessage<CreateEnclaveResponse> | undefined): boolean;
}

/**
 * @generated from message engine_api.EnclaveAPIContainerInfo
 */
export declare class EnclaveAPIContainerInfo extends Message<EnclaveAPIContainerInfo> {
  /**
   * The container engine ID of the API container
   *
   * @generated from field: string container_id = 1;
   */
  containerId: string;

  /**
   * The IP inside the enclave network of the API container (i.e. how services inside the network can reach the API container)
   *
   * @generated from field: string ip_inside_enclave = 2;
   */
  ipInsideEnclave: string;

  /**
   * The grpc port inside the enclave network that the API container is listening on
   *
   * @generated from field: uint32 grpc_port_inside_enclave = 3;
   */
  grpcPortInsideEnclave: number;

  /**
   * this is the bridge ip address that gets assigned to api container
   *
   * @generated from field: string bridge_ip_address = 6;
   */
  bridgeIpAddress: string;

  constructor(data?: PartialMessage<EnclaveAPIContainerInfo>);

  static readonly runtime: typeof proto3;
  static readonly typeName = "engine_api.EnclaveAPIContainerInfo";
  static readonly fields: FieldList;

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): EnclaveAPIContainerInfo;

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): EnclaveAPIContainerInfo;

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): EnclaveAPIContainerInfo;

  static equals(a: EnclaveAPIContainerInfo | PlainMessage<EnclaveAPIContainerInfo> | undefined, b: EnclaveAPIContainerInfo | PlainMessage<EnclaveAPIContainerInfo> | undefined): boolean;
}

/**
 * Will only be present if the API container is running
 *
 * @generated from message engine_api.EnclaveAPIContainerHostMachineInfo
 */
export declare class EnclaveAPIContainerHostMachineInfo extends Message<EnclaveAPIContainerHostMachineInfo> {
  /**
   * The interface IP on the container engine host machine where the API container can be reached
   *
   * @generated from field: string ip_on_host_machine = 4;
   */
  ipOnHostMachine: string;

  /**
   * The grpc port on the container engine host machine where the API container can be reached
   *
   * @generated from field: uint32 grpc_port_on_host_machine = 5;
   */
  grpcPortOnHostMachine: number;

  constructor(data?: PartialMessage<EnclaveAPIContainerHostMachineInfo>);

  static readonly runtime: typeof proto3;
  static readonly typeName = "engine_api.EnclaveAPIContainerHostMachineInfo";
  static readonly fields: FieldList;

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): EnclaveAPIContainerHostMachineInfo;

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): EnclaveAPIContainerHostMachineInfo;

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): EnclaveAPIContainerHostMachineInfo;

  static equals(a: EnclaveAPIContainerHostMachineInfo | PlainMessage<EnclaveAPIContainerHostMachineInfo> | undefined, b: EnclaveAPIContainerHostMachineInfo | PlainMessage<EnclaveAPIContainerHostMachineInfo> | undefined): boolean;
}

/**
 * Enclaves are defined by a network in the container system, which is why there's a bunch of network information here
 *
 * @generated from message engine_api.EnclaveInfo
 */
export declare class EnclaveInfo extends Message<EnclaveInfo> {
  /**
   * UUID of the enclave
   *
   * @generated from field: string enclave_uuid = 1;
   */
  enclaveUuid: string;

  /**
   * Name of the enclave
   *
   * @generated from field: string name = 2;
   */
  name: string;

  /**
   * The shortened uuid of the enclave
   *
   * @generated from field: string shortened_uuid = 3;
   */
  shortenedUuid: string;

  /**
   * State of all containers in the enclave
   *
   * @generated from field: engine_api.EnclaveContainersStatus containers_status = 4;
   */
  containersStatus: EnclaveContainersStatus;

  /**
   * State specifically of the API container
   *
   * @generated from field: engine_api.EnclaveAPIContainerStatus api_container_status = 5;
   */
  apiContainerStatus: EnclaveAPIContainerStatus;

  /**
   * NOTE: Will not be present if the API container status is "NONEXISTENT"!!
   *
   * @generated from field: engine_api.EnclaveAPIContainerInfo api_container_info = 6;
   */
  apiContainerInfo?: EnclaveAPIContainerInfo;

  /**
   * NOTE: Will not be present if the API container status is not "RUNNING"!!
   *
   * @generated from field: engine_api.EnclaveAPIContainerHostMachineInfo api_container_host_machine_info = 7;
   */
  apiContainerHostMachineInfo?: EnclaveAPIContainerHostMachineInfo;

  /**
   * The enclave's creation time
   *
   * @generated from field: google.protobuf.Timestamp creation_time = 8;
   */
  creationTime?: Timestamp;

  /**
   * @generated from field: engine_api.EnclaveMode mode = 9;
   */
  mode: EnclaveMode;

  constructor(data?: PartialMessage<EnclaveInfo>);

  static readonly runtime: typeof proto3;
  static readonly typeName = "engine_api.EnclaveInfo";
  static readonly fields: FieldList;

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): EnclaveInfo;

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): EnclaveInfo;

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): EnclaveInfo;

  static equals(a: EnclaveInfo | PlainMessage<EnclaveInfo> | undefined, b: EnclaveInfo | PlainMessage<EnclaveInfo> | undefined): boolean;
}

/**
 * @generated from message engine_api.GetEnclavesResponse
 */
export declare class GetEnclavesResponse extends Message<GetEnclavesResponse> {
  /**
   * Mapping of enclave_uuid -> info_about_enclave
   *
   * @generated from field: map<string, engine_api.EnclaveInfo> enclave_info = 1;
   */
  enclaveInfo: { [key: string]: EnclaveInfo };

  constructor(data?: PartialMessage<GetEnclavesResponse>);

  static readonly runtime: typeof proto3;
  static readonly typeName = "engine_api.GetEnclavesResponse";
  static readonly fields: FieldList;

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetEnclavesResponse;

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetEnclavesResponse;

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetEnclavesResponse;

  static equals(a: GetEnclavesResponse | PlainMessage<GetEnclavesResponse> | undefined, b: GetEnclavesResponse | PlainMessage<GetEnclavesResponse> | undefined): boolean;
}

/**
 * An enclave identifier is a collection of uuid, name and shortened uuid
 *
 * @generated from message engine_api.EnclaveIdentifiers
 */
export declare class EnclaveIdentifiers extends Message<EnclaveIdentifiers> {
  /**
   * UUID of the enclave
   *
   * @generated from field: string enclave_uuid = 1;
   */
  enclaveUuid: string;

  /**
   * Name of the enclave
   *
   * @generated from field: string name = 2;
   */
  name: string;

  /**
   * The shortened uuid of the enclave
   *
   * @generated from field: string shortened_uuid = 3;
   */
  shortenedUuid: string;

  constructor(data?: PartialMessage<EnclaveIdentifiers>);

  static readonly runtime: typeof proto3;
  static readonly typeName = "engine_api.EnclaveIdentifiers";
  static readonly fields: FieldList;

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): EnclaveIdentifiers;

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): EnclaveIdentifiers;

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): EnclaveIdentifiers;

  static equals(a: EnclaveIdentifiers | PlainMessage<EnclaveIdentifiers> | undefined, b: EnclaveIdentifiers | PlainMessage<EnclaveIdentifiers> | undefined): boolean;
}

/**
 * @generated from message engine_api.GetExistingAndHistoricalEnclaveIdentifiersResponse
 */
export declare class GetExistingAndHistoricalEnclaveIdentifiersResponse extends Message<GetExistingAndHistoricalEnclaveIdentifiersResponse> {
  /**
   * @generated from field: repeated engine_api.EnclaveIdentifiers allIdentifiers = 1;
   */
  allIdentifiers: EnclaveIdentifiers[];

  constructor(data?: PartialMessage<GetExistingAndHistoricalEnclaveIdentifiersResponse>);

  static readonly runtime: typeof proto3;
  static readonly typeName = "engine_api.GetExistingAndHistoricalEnclaveIdentifiersResponse";
  static readonly fields: FieldList;

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetExistingAndHistoricalEnclaveIdentifiersResponse;

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetExistingAndHistoricalEnclaveIdentifiersResponse;

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetExistingAndHistoricalEnclaveIdentifiersResponse;

  static equals(a: GetExistingAndHistoricalEnclaveIdentifiersResponse | PlainMessage<GetExistingAndHistoricalEnclaveIdentifiersResponse> | undefined, b: GetExistingAndHistoricalEnclaveIdentifiersResponse | PlainMessage<GetExistingAndHistoricalEnclaveIdentifiersResponse> | undefined): boolean;
}

/**
 * ==============================================================================================
 *                                       Stop Enclave
 * ==============================================================================================
 *
 * @generated from message engine_api.StopEnclaveArgs
 */
export declare class StopEnclaveArgs extends Message<StopEnclaveArgs> {
  /**
   * The identifier(uuid, shortened uuid, name) of the Kurtosis enclave to stop
   *
   * @generated from field: string enclave_identifier = 1;
   */
  enclaveIdentifier: string;

  constructor(data?: PartialMessage<StopEnclaveArgs>);

  static readonly runtime: typeof proto3;
  static readonly typeName = "engine_api.StopEnclaveArgs";
  static readonly fields: FieldList;

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): StopEnclaveArgs;

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): StopEnclaveArgs;

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): StopEnclaveArgs;

  static equals(a: StopEnclaveArgs | PlainMessage<StopEnclaveArgs> | undefined, b: StopEnclaveArgs | PlainMessage<StopEnclaveArgs> | undefined): boolean;
}

/**
 * ==============================================================================================
 *                                       Destroy Enclave
 * ==============================================================================================
 *
 * @generated from message engine_api.DestroyEnclaveArgs
 */
export declare class DestroyEnclaveArgs extends Message<DestroyEnclaveArgs> {
  /**
   * The identifier(uuid, shortened uuid, name) of the Kurtosis enclave to destroy
   *
   * @generated from field: string enclave_identifier = 1;
   */
  enclaveIdentifier: string;

  constructor(data?: PartialMessage<DestroyEnclaveArgs>);

  static readonly runtime: typeof proto3;
  static readonly typeName = "engine_api.DestroyEnclaveArgs";
  static readonly fields: FieldList;

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): DestroyEnclaveArgs;

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): DestroyEnclaveArgs;

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): DestroyEnclaveArgs;

  static equals(a: DestroyEnclaveArgs | PlainMessage<DestroyEnclaveArgs> | undefined, b: DestroyEnclaveArgs | PlainMessage<DestroyEnclaveArgs> | undefined): boolean;
}

/**
 * ==============================================================================================
 *                                       Create Enclave
 * ==============================================================================================
 *
 * @generated from message engine_api.CleanArgs
 */
export declare class CleanArgs extends Message<CleanArgs> {
  /**
   * If true, It will clean even the running enclaves
   *
   * @generated from field: bool should_clean_all = 1;
   */
  shouldCleanAll: boolean;

  constructor(data?: PartialMessage<CleanArgs>);

  static readonly runtime: typeof proto3;
  static readonly typeName = "engine_api.CleanArgs";
  static readonly fields: FieldList;

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): CleanArgs;

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): CleanArgs;

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): CleanArgs;

  static equals(a: CleanArgs | PlainMessage<CleanArgs> | undefined, b: CleanArgs | PlainMessage<CleanArgs> | undefined): boolean;
}

/**
 * @generated from message engine_api.EnclaveNameAndUuid
 */
export declare class EnclaveNameAndUuid extends Message<EnclaveNameAndUuid> {
  /**
   * @generated from field: string name = 1;
   */
  name: string;

  /**
   * @generated from field: string uuid = 2;
   */
  uuid: string;

  constructor(data?: PartialMessage<EnclaveNameAndUuid>);

  static readonly runtime: typeof proto3;
  static readonly typeName = "engine_api.EnclaveNameAndUuid";
  static readonly fields: FieldList;

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): EnclaveNameAndUuid;

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): EnclaveNameAndUuid;

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): EnclaveNameAndUuid;

  static equals(a: EnclaveNameAndUuid | PlainMessage<EnclaveNameAndUuid> | undefined, b: EnclaveNameAndUuid | PlainMessage<EnclaveNameAndUuid> | undefined): boolean;
}

/**
 * @generated from message engine_api.CleanResponse
 */
export declare class CleanResponse extends Message<CleanResponse> {
  /**
   * removed enclave name and uuids
   *
   * @generated from field: repeated engine_api.EnclaveNameAndUuid removed_enclave_name_and_uuids = 1;
   */
  removedEnclaveNameAndUuids: EnclaveNameAndUuid[];

  constructor(data?: PartialMessage<CleanResponse>);

  static readonly runtime: typeof proto3;
  static readonly typeName = "engine_api.CleanResponse";
  static readonly fields: FieldList;

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): CleanResponse;

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): CleanResponse;

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): CleanResponse;

  static equals(a: CleanResponse | PlainMessage<CleanResponse> | undefined, b: CleanResponse | PlainMessage<CleanResponse> | undefined): boolean;
}

/**
 * ==============================================================================================
 *                                   Get User Service Logs
 * ==============================================================================================
 *
 * @generated from message engine_api.GetServiceLogsArgs
 */
export declare class GetServiceLogsArgs extends Message<GetServiceLogsArgs> {
  /**
   * The identifier of the user service's Kurtosis Enclave
   *
   * @generated from field: string enclave_identifier = 1;
   */
  enclaveIdentifier: string;

  /**
   * "Set" of service UUIDs in the enclave
   *
   * @generated from field: map<string, bool> service_uuid_set = 2;
   */
  serviceUuidSet: { [key: string]: boolean };

  /**
   * If true, It will follow the container logs
   *
   * @generated from field: bool follow_logs = 3;
   */
  followLogs: boolean;

  /**
   * The conjunctive log lines filters, the first filter is applied over the found log lines, the second filter is applied over the filter one result and so on (like grep)
   *
   * @generated from field: repeated engine_api.LogLineFilter conjunctive_filters = 4;
   */
  conjunctiveFilters: LogLineFilter[];

  /**
   * If true, return all log lines
   *
   * @generated from field: bool return_all_logs = 5;
   */
  returnAllLogs: boolean;

  /**
   * If [return_all_logs] is false, return [num_log_lines]
   *
   * @generated from field: uint32 num_log_lines = 6;
   */
  numLogLines: number;

  constructor(data?: PartialMessage<GetServiceLogsArgs>);

  static readonly runtime: typeof proto3;
  static readonly typeName = "engine_api.GetServiceLogsArgs";
  static readonly fields: FieldList;

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetServiceLogsArgs;

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetServiceLogsArgs;

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetServiceLogsArgs;

  static equals(a: GetServiceLogsArgs | PlainMessage<GetServiceLogsArgs> | undefined, b: GetServiceLogsArgs | PlainMessage<GetServiceLogsArgs> | undefined): boolean;
}

/**
 * @generated from message engine_api.GetServiceLogsResponse
 */
export declare class GetServiceLogsResponse extends Message<GetServiceLogsResponse> {
  /**
   * The service log lines grouped by service UUIDs and ordered in forward direction (oldest log line is the first element)
   *
   * @generated from field: map<string, engine_api.LogLine> service_logs_by_service_uuid = 1;
   */
  serviceLogsByServiceUuid: { [key: string]: LogLine };

  /**
   * A set of service GUIDs requested by the user that were not found in the logs database, could be related that users send
   * a wrong GUID or a right GUID for a service that has not sent any logs so far
   *
   * @generated from field: map<string, bool> not_found_service_uuid_set = 2;
   */
  notFoundServiceUuidSet: { [key: string]: boolean };

  constructor(data?: PartialMessage<GetServiceLogsResponse>);

  static readonly runtime: typeof proto3;
  static readonly typeName = "engine_api.GetServiceLogsResponse";
  static readonly fields: FieldList;

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetServiceLogsResponse;

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetServiceLogsResponse;

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetServiceLogsResponse;

  static equals(a: GetServiceLogsResponse | PlainMessage<GetServiceLogsResponse> | undefined, b: GetServiceLogsResponse | PlainMessage<GetServiceLogsResponse> | undefined): boolean;
}

/**
 * TODO add timestamp as well, for when we do timestamp-handling on the client side
 *
 * @generated from message engine_api.LogLine
 */
export declare class LogLine extends Message<LogLine> {
  /**
   * @generated from field: repeated string line = 1;
   */
  line: string[];

  constructor(data?: PartialMessage<LogLine>);

  static readonly runtime: typeof proto3;
  static readonly typeName = "engine_api.LogLine";
  static readonly fields: FieldList;

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): LogLine;

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): LogLine;

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): LogLine;

  static equals(a: LogLine | PlainMessage<LogLine> | undefined, b: LogLine | PlainMessage<LogLine> | undefined): boolean;
}

/**
 * @generated from message engine_api.LogLineFilter
 */
export declare class LogLineFilter extends Message<LogLineFilter> {
  /**
   * @generated from field: engine_api.LogLineOperator operator = 1;
   */
  operator: LogLineOperator;

  /**
   * @generated from field: string text_pattern = 2;
   */
  textPattern: string;

  constructor(data?: PartialMessage<LogLineFilter>);

  static readonly runtime: typeof proto3;
  static readonly typeName = "engine_api.LogLineFilter";
  static readonly fields: FieldList;

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): LogLineFilter;

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): LogLineFilter;

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): LogLineFilter;

  static equals(a: LogLineFilter | PlainMessage<LogLineFilter> | undefined, b: LogLineFilter | PlainMessage<LogLineFilter> | undefined): boolean;
}

