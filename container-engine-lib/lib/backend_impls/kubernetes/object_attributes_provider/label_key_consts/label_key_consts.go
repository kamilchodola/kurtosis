package label_key_consts

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_key"
)

const (
	// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! DO NOT CHANGE THESE VALUES !!!!!!!!!!!!!!!!!!!!!!!!!!!!!
	// If these value change, it will lead to the Kurtosis engine losing track of old containers
	//  which will cause a resource leak on the user's system!
	//
	//   If you add new immutable values to this section, MAKE SURE TO UPDATE THE UNIT TEST!
	//
	labelKeyPrefixStr       = "kurtosistech.com/"
	appIdLabelKeyStr        = labelKeyPrefixStr + "app-id"
	resourceTypeLabelKeyStr = labelKeyPrefixStr + "resource-type"
	// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! DO NOT CHANGE THESE VALUES !!!!!!!!!!!!!!!!!!!!!!!!!!!!!

	// A label to identify a Kurtosis resource (e.g. network, container, etc.) by its id
	idLabelKeyStr = labelKeyPrefixStr + "id"

	// Used for things like service GUID, module GUID, etc.
	guidLabelKeyStr = labelKeyPrefixStr + "guid"

	portSpecsLabelKeyStr = labelKeyPrefixStr + "ports"

	enclaveIdLabelKeyStr = labelKeyPrefixStr + "enclave-id"

	isNetworkPartitioningEnabledKeyStr = labelKeyPrefixStr + "is-network-partitioning-enabled"

	// TODO Remove this and instead use container domain names so we're not dependent on static IPs!
	privateIpAddrLabelKeyStr = labelKeyPrefixStr + "private-ip"
)

// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! DO NOT CHANGE THESE VALUES !!!!!!!!!!!!!!!!!!!!!!!!!!!!!
// If these value change, it will lead to the Kurtosis engine losing track of old resources
//  which will cause a resource leak on the user's cluster!
//
//   If you add new immutable values to this section, MAKE SURE TO UPDATE THE UNIT TEST!
//
var AppIDLabelKey = kubernetes_label_key.MustCreateNewKubernetesLabelKey(appIdLabelKeyStr)
var ResourceTypeLabelKey = kubernetes_label_key.MustCreateNewKubernetesLabelKey(resourceTypeLabelKeyStr)

// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! DO NOT CHANGE THESE VALUES !!!!!!!!!!!!!!!!!!!!!!!!!!!!!

var IDLabelKey = kubernetes_label_key.MustCreateNewKubernetesLabelKey(idLabelKeyStr)
var GUIDLabelKey = kubernetes_label_key.MustCreateNewKubernetesLabelKey(guidLabelKeyStr)
var PortSpecsLabelKey = kubernetes_label_key.MustCreateNewKubernetesLabelKey(portSpecsLabelKeyStr)
var EnclaveIDLabelKey = kubernetes_label_key.MustCreateNewKubernetesLabelKey(enclaveIdLabelKeyStr)
var IsNetworkPartitioningEnabledLabelKey = kubernetes_label_key.MustCreateNewKubernetesLabelKey(isNetworkPartitioningEnabledKeyStr)
var PrivateIPLabelKey = kubernetes_label_key.MustCreateNewKubernetesLabelKey(privateIpAddrLabelKeyStr)
