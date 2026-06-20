package dbricksLks

type AuthType string

const (
	AuthTypeAzureClientSecret AuthType = "azure-client-secret"
	AuthTypeGoogleId          AuthType = "google-id"
	AuthTypeGoogleCredentials AuthType = "google-credentials"
	AuthTypeAzureCli          AuthType = "azure-cli,azure-msi"
	AuthTypeAzureMsi          AuthType = "azure-msi"
	AuthTypeOAuthM2M          AuthType = "oauth-m2m"
	AuthTypeDatabricksCli     AuthType = "databricks-cli"
)

type Resource struct {
	Id   string `yaml:"id,omitempty" json:"id,omitempty" mapstructure:"id,omitempty"`
	Name string `yaml:"name,omitempty" json:"name,omitempty" mapstructure:"name,omitempty"`
}

type ServicePrincipal struct {
	TenantID     string `yaml:"tenant-id,omitempty" json:"tenant-id,omitempty" mapstructure:"tenant-id,omitempty"`
	ClientID     string `yaml:"client-id,omitempty" json:"client-id,omitempty" mapstructure:"client-id,omitempty"`
	ClientSecret string `yaml:"client-secret,omitempty" json:"client-secret,omitempty" mapstructure:"client-secret,omitempty"`
}

type Config struct {
	Name                string            `yaml:"name,omitempty" json:"name,omitempty" mapstructure:"name,omitempty"`
	Host                string            `yaml:"host,omitempty" json:"host,omitempty" mapstructure:"host,omitempty"`
	WorkspaceResourceID string            `yaml:"workspace-resource-id,omitempty" json:"workspace-resource-id,omitempty" mapstructure:"workspace-resource-id,omitempty"`
	AuthType            AuthType          `yaml:"auth-type,omitempty" json:"auth-type,omitempty" mapstructure:"auth-type,omitempty"`
	ServicePrincipal    *ServicePrincipal `yaml:"service-principal,omitempty" json:"service-principal,omitempty" mapstructure:"service-principal,omitempty"`
	WarehouseID         string            `yaml:"warehouse-id,omitempty" json:"warehouse-id,omitempty" mapstructure:"warehouse-id,omitempty"`
	Resources           []Resource        `yaml:"catalog-resources,omitempty" json:"catalog-resources,omitempty" mapstructure:"catalog-resources,omitempty"`
}

type Option func(cfg *Config)

func WithName(k string) Option {
	return func(cfg *Config) {
		cfg.Name = k
	}
}
