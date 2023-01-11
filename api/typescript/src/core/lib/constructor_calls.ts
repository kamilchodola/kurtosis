/*
 * Copyright (c) 2022 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

import * as jspb from "google-protobuf";
import {
    ExecCommandArgs,
    GetServicesArgs,
    PartitionServices,
    PartitionConnections,
    PartitionConnectionInfo,
    RemoveServiceArgs,
    RepartitionArgs,
    WaitForHttpGetEndpointAvailabilityArgs,
    WaitForHttpPostEndpointAvailabilityArgs,
    Port,
    StoreWebFilesArtifactArgs,
    StoreFilesArtifactFromServiceArgs,
    UploadFilesArtifactArgs,
    PauseServiceArgs,
    UnpauseServiceArgs,
    ServiceInfo,
    ServiceConfig,
    RemoveServiceResponse,
    GetServicesResponse, StartServicesArgs,
    RenderTemplatesToFilesArtifactArgs, DownloadFilesArtifactArgs,
} from '../kurtosis_core_rpc_api_bindings/api_container_service_pb';
import { ServiceID } from './services/service';
import TemplateAndData = RenderTemplatesToFilesArtifactArgs.TemplateAndData;

// ==============================================================================================
//                           Shared Objects (Used By Multiple Endpoints)
// ==============================================================================================
export function newPort(number: number, transportProtocol: Port.TransportProtocol, maybeApplicationProtocol?: string) {
    const result: Port = new Port();
    result.setNumber(number);
    result.setTransportProtocol(transportProtocol);
    if (maybeApplicationProtocol) {
        result.setMaybeApplicationProtocol(maybeApplicationProtocol)
    }
    return result;
}

export function newServiceConfig(
    containerImageName : string,
    privatePorts : Map<string, Port>,
    publicPorts : Map<string, Port>,
    entrypointOverrideArgs: string[],
    cmdOverrideArgs: string[],
    environmentVariableOverrides : Map<string, string>,
    filesArtifactMountDirpaths : Map<string, string>,
    cpuAllocationMillicpus : number,
    memoryAllocationMegabytes : number,
    privateIPAddrPlaceholder : string,
    subnetwork : string,
) {
    const result : ServiceConfig = new ServiceConfig();
    result.setContainerImageName(containerImageName);
    const usedPortsMap: jspb.Map<string, Port> = result.getPrivatePortsMap();
    for (const [portId, portSpec] of privatePorts) {
        usedPortsMap.set(portId, portSpec);
    }
    //TODO this is a huge hack to temporarily enable static ports for NEAR until we have a more productized solution
    const publicPortsMap: jspb.Map<string, Port> = result.getPublicPortsMap();
    for (const [portId, portSpec] of publicPorts) {
        publicPortsMap.set(portId, portSpec);
    }
    //TODO finish the hack
    const entrypointArgsArray: string[] = result.getEntrypointArgsList();
    for (const entryPoint of entrypointOverrideArgs) {
        entrypointArgsArray.push(entryPoint);
    }
    const cmdArgsArray: string[] = result.getCmdArgsList();
    for (const cmdArg of cmdOverrideArgs) {
        cmdArgsArray.push(cmdArg);
    }
    const envVarArray: jspb.Map<string, string> = result.getEnvVarsMap();
    for (const [name, value] of environmentVariableOverrides) {
        envVarArray.set(name, value);
    }
    const filesArtifactMountDirpathsMap: jspb.Map<string, string> = result.getFilesArtifactMountpointsMap();
    for (const [artifactId, mountDirpath] of filesArtifactMountDirpaths) {
        filesArtifactMountDirpathsMap.set(artifactId, mountDirpath);
    }
    result.setCpuAllocationMillicpus(cpuAllocationMillicpus);
    result.setMemoryAllocationMegabytes(memoryAllocationMegabytes);
    result.setPrivateIpAddrPlaceholder(privateIPAddrPlaceholder);
    result.setSubnetwork(subnetwork);
    return result;
}


// ==============================================================================================
//                                        Start Service
// ==============================================================================================
export function newStartServicesArgs(serviceConfigs : Map<ServiceID, ServiceConfig>) : StartServicesArgs {
    const result : StartServicesArgs = new StartServicesArgs();
    const serviceIdsToConfigs : jspb.Map<string, ServiceConfig> = result.getServiceIdsToConfigsMap();
    for (const [serviceId, serviceConfig] of serviceConfigs) {
        serviceIdsToConfigs.set(String(serviceId), serviceConfig);
    }
    return result;
}

// ==============================================================================================
//                                       Get Services
// ==============================================================================================
export function newGetServicesArgs(serviceIds: Map<string, boolean>): GetServicesArgs{
    const result: GetServicesArgs = new GetServicesArgs();
    const resultServiceIdMap: jspb.Map<string, boolean> = result.getServiceIdsMap()
    for (const [serviceId, booleanVal] of serviceIds) {
        resultServiceIdMap.set(serviceId, booleanVal);
    }

    return result;
}

export function newGetServicesResponse(serviceInfoMap: Map<string,ServiceInfo>): GetServicesResponse{
    const result: GetServicesResponse = new GetServicesResponse();
    const resultServiceMap: jspb.Map<string,ServiceInfo> = result.getServiceInfoMap()
    for (const [serviceId, serviceInfo] of serviceInfoMap) {
        resultServiceMap.set(serviceId, serviceInfo)
    }

    return result
}

export function newServiceInfo(
    serviceGuid: string,
    privateIpAddr: string,
    privatePorts: Map<string, Port>,
    maybePublicIpAddr: string,
    maybePublicPorts: Map<string, Port>,
): ServiceInfo {
    const result: ServiceInfo = new ServiceInfo();
    result.setServiceGuid(serviceGuid)
    result.setMaybePublicIpAddr(maybePublicIpAddr)
    result.setPrivateIpAddr(privateIpAddr)

    const privatePortsMap: jspb.Map<string, Port> = result.getPrivatePortsMap()
    for (const [portName, privatePort] of privatePorts.entries()) {
        privatePortsMap.set(portName, privatePort)
    }
    const maybePublicPortsMap: jspb.Map<string, Port> = result.getMaybePublicPortsMap()
    for (const [portName, publicPort] of maybePublicPorts.entries()) {
        maybePublicPortsMap.set(portName, publicPort)
    }

    return result
}


// ==============================================================================================
//                                        Remove Service
// ==============================================================================================
export function newRemoveServiceArgs(serviceId: ServiceID): RemoveServiceArgs {
    const result: RemoveServiceArgs = new RemoveServiceArgs();
    result.setServiceId(serviceId);

    return result;
}

export function newRemoveServiceResponse(serviceGuid: string): RemoveServiceResponse {
    const result: RemoveServiceResponse = new RemoveServiceResponse();
    result.setServiceGuid(serviceGuid)
    return result
}


// ==============================================================================================
//                                          Repartition
// ==============================================================================================
export function newRepartitionArgs(
        partitionServices: Map<string, PartitionServices>, 
        partitionConns: Map<string, PartitionConnections>,
        defaultConnection: PartitionConnectionInfo): RepartitionArgs {
    const result: RepartitionArgs = new RepartitionArgs();
    const partitionServicesMap: jspb.Map<string, PartitionServices> = result.getPartitionServicesMap();
    for (const [partitionServiceId, partitionId] of partitionServices.entries()) {
        partitionServicesMap.set(partitionServiceId, partitionId);
    };
    const partitionConnsMap: jspb.Map<string, PartitionConnections> = result.getPartitionConnectionsMap();
    for (const [partitionConnId, partitionConn] of partitionConns.entries()) {
        partitionConnsMap.set(partitionConnId, partitionConn);
    };
    result.setDefaultConnection(defaultConnection);

    return result;
}

export function newPartitionServices(serviceIdStrSet: Set<string>): PartitionServices{
    const result: PartitionServices = new PartitionServices();
    const partitionServicesMap: jspb.Map<string, boolean> = result.getServiceIdSetMap();
    for (const serviceIdStr of serviceIdStrSet) {
        partitionServicesMap.set(serviceIdStr, true);
    }

    return result;
}


export function newPartitionConnections(allConnectionInfo: Map<string, PartitionConnectionInfo>): PartitionConnections {
    const result: PartitionConnections = new PartitionConnections();
    const partitionsMap: jspb.Map<string, PartitionConnectionInfo> = result.getConnectionInfoMap();
    for (const [partitionId, connectionInfo] of allConnectionInfo.entries()) {
        partitionsMap.set(partitionId, connectionInfo);
    }

    return result;
}

export function newPartitionConnectionInfo(packetLossPercentage: number): PartitionConnectionInfo {
    const partitionConnectionInfo: PartitionConnectionInfo = new PartitionConnectionInfo();
    partitionConnectionInfo.setPacketLossPercentage(packetLossPercentage);
    return partitionConnectionInfo;
}


// ==============================================================================================
//                                          Exec Command
// ==============================================================================================
export function newExecCommandArgs(serviceId: ServiceID, command: string[]): ExecCommandArgs {
    const result: ExecCommandArgs = new ExecCommandArgs();
    result.setServiceId(serviceId);
    result.setCommandArgsList(command);

    return result;
}

// ==============================================================================================
//                                          Pause/Unpause Service
// ==============================================================================================
export function newPauseServiceArgs(serviceId: ServiceID): PauseServiceArgs {
    const result: PauseServiceArgs = new PauseServiceArgs();
    result.setServiceId(serviceId);

    return result;
}

export function newUnpauseServiceArgs(serviceId: ServiceID): UnpauseServiceArgs {
    const result: UnpauseServiceArgs = new UnpauseServiceArgs();
    result.setServiceId(serviceId);

    return result;
}


// ==============================================================================================
//                           Wait For Http Get Endpoint Availability
// ==============================================================================================
export function newWaitForHttpGetEndpointAvailabilityArgs(
        serviceId: ServiceID,
        port: number, 
        path: string,
        initialDelayMilliseconds: number, 
        retries: number, 
        retriesDelayMilliseconds: number, 
        bodyText: string): WaitForHttpGetEndpointAvailabilityArgs {
    const result: WaitForHttpGetEndpointAvailabilityArgs = new WaitForHttpGetEndpointAvailabilityArgs();
    result.setServiceId(String(serviceId));
    result.setPort(port);
    result.setPath(path);
    result.setInitialDelayMilliseconds(initialDelayMilliseconds);
    result.setRetries(retries);
    result.setRetriesDelayMilliseconds(retriesDelayMilliseconds);
    result.setBodyText(bodyText);

    return result;
}


// ==============================================================================================
//                           Wait For Http Post Endpoint Availability
// ==============================================================================================
export function newWaitForHttpPostEndpointAvailabilityArgs(
        serviceId: ServiceID,
        port: number, 
        path: string,
        requestBody: string,
        initialDelayMilliseconds: number, 
        retries: number, 
        retriesDelayMilliseconds: number, 
        bodyText: string): WaitForHttpPostEndpointAvailabilityArgs {
    const result: WaitForHttpPostEndpointAvailabilityArgs = new WaitForHttpPostEndpointAvailabilityArgs();
    result.setServiceId(String(serviceId));
    result.setPort(port);
    result.setPath(path);
    result.setRequestBody(requestBody)
    result.setInitialDelayMilliseconds(initialDelayMilliseconds);
    result.setRetries(retries);
    result.setRetriesDelayMilliseconds(retriesDelayMilliseconds);
    result.setBodyText(bodyText);

    return result;
}

// ==============================================================================================
//                                     Store Web Files Files
// ==============================================================================================
export function newStoreWebFilesArtifactArgs(url: string, name: string): StoreWebFilesArtifactArgs {
    const result: StoreWebFilesArtifactArgs = new StoreWebFilesArtifactArgs();
    result.setUrl(url);
    result.setName(name);
    return result;
}

// ==============================================================================================
//                                     Download Files
// ==============================================================================================
export function newDownloadFilesArtifactArgs(identifier: string): DownloadFilesArtifactArgs {
    const result: DownloadFilesArtifactArgs = new DownloadFilesArtifactArgs();
    result.setIdentifier(identifier);
    return result;
}

// ==============================================================================================
//                             Store Files Artifact From Service
// ==============================================================================================
export function newStoreFilesArtifactFromServiceArgs(serviceId: string, sourcePath: string): StoreFilesArtifactFromServiceArgs {
    const result: StoreFilesArtifactFromServiceArgs = new StoreFilesArtifactFromServiceArgs();
    result.setServiceId(serviceId)
    result.setSourcePath(sourcePath)
    return result;
}

// ==============================================================================================
//                                      Upload Files
// ==============================================================================================
export function newUploadFilesArtifactArgs(data: Uint8Array, name: string) : UploadFilesArtifactArgs {
    const result: UploadFilesArtifactArgs = new UploadFilesArtifactArgs()
    result.setData(data)
    result.setName(name)
    return result
}

// ==============================================================================================
//                                      Render Templates
// ==============================================================================================
export function newTemplateAndData(template: string, templateData: string) : TemplateAndData {
    const templateAndData : TemplateAndData = new TemplateAndData()
    templateAndData.setDataAsJson(templateData)
    templateAndData.setTemplate(template)
    return templateAndData
}

export function newRenderTemplatesToFilesArtifactArgs() : RenderTemplatesToFilesArtifactArgs {
    const renderTemplatesToFilesArtifactArgs : RenderTemplatesToFilesArtifactArgs = new RenderTemplatesToFilesArtifactArgs()
    return renderTemplatesToFilesArtifactArgs
}
