package v1

import (
	"github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_config/config_version"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	kurtosisConfigV1DockerType     = "docker"
	kurtosisConfigV1KubernetesType = "kubernetes"
)

// NOTE: All new YAML property names here should be kebab-case because
//a) it's easier to read b) it's easier to write
//c) it's consistent with previous properties and changing the format of
//an already-written config file is very difficult

const (
	versionNumber = config_version.ConfigVersion_v1
	defaultDockerClusterName = "docker"
	defaultMinikubeClusterName = "minikube"
	defaultMinikubeStorageClass = "standard"
	defaultMinikubeGigabytesPerEnclave = 2
)

type KurtosisConfigV1 struct {
	//We set public fields because YAML marshalling needs it on this way
	//All fields should be pointers, that way we can enforce required fields
	//by detecting nil pointers.
	ConfigVersion *config_version.ConfigVersion `yaml:"config-version,omitempty"`
	ShouldSendMetrics *bool                         `yaml:"should-send-metrics,omitempty"`
	KurtosisClusters *map[string]*KurtosisClusterV1 `yaml:"kurtosis-clusters,omitempty"`
}

type KurtosisClusterV1 struct {
	Type *string                      `yaml:"type,omitempty"`
	// If we ever get another type of cluster that has configuration, this will need to be polymorphically deserialized
	Config *KubernetesClusterConfigV1 `yaml:"config,omitempty"`
}

type KubernetesClusterConfigV1 struct {
	KubernetesClusterName *string `yaml:"kubernetes-cluster-name,omitempty"`
	StorageClass *string `yaml:"storage-class,omitempty"`
	EnclaveSizeInGigabytes *int `yaml:"enclave-size-in-gigabytes,omitempty"`
}

func NewDefaultKurtosisConfigV1() *KurtosisConfigV1 {
	version := versionNumber
	dockerClusterConfig := getDefaultDockerKurtosisClusterConfig()
	kurtosisClusters := map[string]*KurtosisClusterV1{
		defaultDockerClusterName:   dockerClusterConfig,
	}
	return &KurtosisConfigV1{
		ConfigVersion: &version,
		KurtosisClusters: &kurtosisClusters,
		ShouldSendMetrics: nil, // ShouldSendMetrics is nil because it MUST be a user input - can not be a default
	}
}

func (kurtosisConfigV1 *KurtosisConfigV1) Validate() error {
	if kurtosisConfigV1.ShouldSendMetrics == nil {
		return stacktrace.NewError("ShouldSendMetrics field of Kurtosis Config v1 is nil, when it should be true or false.")
	}
	if kurtosisConfigV1.ConfigVersion == nil {
		return stacktrace.NewError("ConfigVersion field of Kurtosis Config v1 is nil, when it should be %d.", versionNumber)
	}
	if *kurtosisConfigV1.ConfigVersion != versionNumber {
		return stacktrace.NewError("ConfigVersion field of Kurtosis Config v1 is %d, when it should be %d.", kurtosisConfigV1.ConfigVersion, versionNumber)
	}
	if kurtosisConfigV1.KurtosisClusters == nil {
		return stacktrace.NewError("KurtosisCluster field of Kurtosis Config v1 is nil, when it should have a map of Kurtosis cluster configurations.")
	}
	if len(*kurtosisConfigV1.KurtosisClusters) == 0 {
		return stacktrace.NewError("KurtosisCluster field of Kurtosis Config v1 has no clusters, when it should have at least one.")
	}
	for clusterId, clusterConfig := range *kurtosisConfigV1.KurtosisClusters {
		if err := clusterConfig.Validate(clusterId); err != nil {
			return stacktrace.Propagate(err, "Failed to validate KurtosisCluster configuration for clusterId '%v'", clusterId)
		}
	}
	return nil
}

func (kurtosisConfigV1 *KurtosisConfigV1) OverlayOverrides(overrides *KurtosisConfigV1) (*KurtosisConfigV1, error) {
	baseKurtosisConfig := kurtosisConfigV1
	if overrides.ShouldSendMetrics != nil {
		baseKurtosisConfig.ShouldSendMetrics = overrides.ShouldSendMetrics
	}
	if overrides.KurtosisClusters != nil {
		baseClusterMap := *baseKurtosisConfig.KurtosisClusters
		for clusterId, clusterConfig := range *overrides.KurtosisClusters {
			baseClusterConfig := baseClusterMap[clusterId]
			if baseClusterConfig != nil {
				overlaidClusterConfig, err := baseClusterConfig.OverlayOverrides(clusterConfig)
				if err != nil {
					return nil, stacktrace.Propagate(err, "Failed to overlay configuration overrides for clusterId '%v'", clusterId)
				}
				baseClusterMap[clusterId] = overlaidClusterConfig
			} else {
				baseClusterMap[clusterId] = clusterConfig
			}
		}
		baseKurtosisConfig.KurtosisClusters = &baseClusterMap
	}
	return baseKurtosisConfig, nil
}

func (kurtosisClusterV1 *KurtosisClusterV1) OverlayOverrides(overrides *KurtosisClusterV1) (*KurtosisClusterV1, error) {
	overlaidKurtosisClusterConfig := kurtosisClusterV1
	if overrides.Type != nil {
		overlaidKurtosisClusterConfig.Type = overrides.Type
	}
	if overrides.Config != nil {
		configOverrides := overlaidKurtosisClusterConfig.Config.OverlayOverrides(overrides.Config)
		overlaidKurtosisClusterConfig.Config = configOverrides
	}
	return overlaidKurtosisClusterConfig, nil
}

func (kubernetesClusterV1 *KubernetesClusterConfigV1) OverlayOverrides(overrides *KubernetesClusterConfigV1) *KubernetesClusterConfigV1 {
	overlaidKubernetesClusterConfig := kubernetesClusterV1
	if overrides.KubernetesClusterName != nil {
		overlaidKubernetesClusterConfig.KubernetesClusterName = overrides.KubernetesClusterName
	}
	if overrides.EnclaveSizeInGigabytes != nil {
		overlaidKubernetesClusterConfig.EnclaveSizeInGigabytes = overrides.EnclaveSizeInGigabytes
	}
	if overrides.StorageClass != nil {
		overlaidKubernetesClusterConfig.StorageClass = overrides.StorageClass
	}
	return overlaidKubernetesClusterConfig
}

func (kurtosisClusterV1 *KurtosisClusterV1) Validate(clusterId string) error {
	clusterConfig := kurtosisClusterV1
	if clusterConfig.Type == nil {
		return stacktrace.NewError("KurtosisCluster '%v' has nil Type field, when it should be'%v' or '%v'",
			clusterId, kurtosisConfigV1DockerType, kurtosisConfigV1KubernetesType)
	}
	if *clusterConfig.Type != kurtosisConfigV1DockerType && *clusterConfig.Type != kurtosisConfigV1KubernetesType {
		return stacktrace.NewError("KurtosisCluster '%v' has Type field '%v', when it should be '%v' or '%v'",
			clusterId, *clusterConfig.Type, kurtosisConfigV1DockerType, kurtosisConfigV1KubernetesType)
	}
	if *clusterConfig.Type == kurtosisConfigV1KubernetesType {
		if clusterConfig.Config == nil {
			return stacktrace.NewError("KurtosisCluster '%v' has Type field '%v' but has no Config field. Config fields are required for type '%v'",
				clusterId, *clusterConfig.Type, kurtosisConfigV1KubernetesType)
		}
		if clusterConfig.Config.KubernetesClusterName == nil {
			return stacktrace.NewError("KurtosisCluster '%v' has Type field '%v' but has no Kubernetes cluster name in its config map.",
				clusterId, *clusterConfig.Type)
		}
		if clusterConfig.Config.StorageClass == nil {
			return stacktrace.NewError("KurtosisCluster '%v' has Type field '%v' but has no StorageClass name in its config map.",
				clusterId, *clusterConfig.Type)
		}
		if clusterConfig.Config.EnclaveSizeInGigabytes == nil {
			return stacktrace.NewError("KurtosisCluster '%v' has Type field '%v' but has no EnclaveSizeInGigabytes specified in its config map.",
				clusterId, *clusterConfig.Type)
		}
	}
	return nil
}

// ===================== HELPERS==============================

func getDefaultDockerKurtosisClusterConfig() *KurtosisClusterV1 {
	clusterType := kurtosisConfigV1DockerType
	return &KurtosisClusterV1{
		Type: &clusterType,
	}
}

func getDefaultMinikubeKurtosisClusterConfig() *KurtosisClusterV1 {
	clusterType := kurtosisConfigV1KubernetesType
	minikubeKubernetesCluster := getDefaultMinikubeKubernetesClusterConfig()
	return &KurtosisClusterV1{
		Type: &clusterType,
		Config: minikubeKubernetesCluster,
	}
}

func getDefaultMinikubeKubernetesClusterConfig() *KubernetesClusterConfigV1 {
	kubernetesClusterName := defaultMinikubeClusterName
	storageClass := defaultMinikubeStorageClass
	gbPerEnclave := defaultMinikubeGigabytesPerEnclave
	clusterConfig := KubernetesClusterConfigV1{
		KubernetesClusterName: &kubernetesClusterName,
		StorageClass: &storageClass,
		EnclaveSizeInGigabytes: &gbPerEnclave,
	}
	return &clusterConfig
}

