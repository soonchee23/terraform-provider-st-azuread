package azuread

import (
	"context"
	"os"

	azureClient "github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	graph "github.com/microsoftgraph/msgraph-sdk-go"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Wrapper of Azuread client
type azureadClients struct {
	graphClient *graph.GraphServiceClient
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ provider.Provider = &azureadProvider{}
)

// New is a helper function to simplify provider server
func New() provider.Provider {
	return &azureadProvider{}
}

// azureadProvider is the provider implementation.
type azureadProvider struct{}

// azureadProviderModel maps provider schema data to a Go type.
type azureadProviderModel struct {
	TenantID     types.String `tfsdk:"tenant_id"`
	ClientID     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
}

// Metadata returns the provider type name.
func (p *azureadProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "st-azuread"
}

// Schema defines the provider-level schema for configuration data.
func (p *azureadProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The AzureAD provider is used to interact with the many resources supported by Microsoft GraphAPI. " +
			"The provider needs to be configured with the proper credentials before it can be used.",
		Attributes: map[string]schema.Attribute{
			"tenant_id": schema.StringAttribute{
				Description: "Tenant ID for MS Graph API. May also be provided via AZURE_TENANT_ID environment variable.",
				Optional:    true,
			},
			"client_id": schema.StringAttribute{
				Description: "Client ID for MS Graph API. May also be provided via AZURE_CLIENT_ID environment variable.",
				Optional:    true,
			},
			"client_secret": schema.StringAttribute{
				Description: "Client Secret for MS Graph API. May also be provided via AZURE_CLIENT_SECRET environment variable.",
				Optional:    true,
			},
		},
	}
}

func (p *azureadProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config azureadProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the attributes,
	// it must be a known value.
	if config.TenantID.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("tenant_id"),
			"Unknown Graph tenant id",
			"The provider cannot create the Graph API client as there is an unknown configuration value for the"+
				"Graph API tenant id. Set the value statically in the configuration, or use the AZURE_TENANT_ID environment variable.",
		)
	}

	if config.ClientID.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("client_id"),
			"Unknown Graph client id",
			"The provider cannot create the Graph API client as there is an unknown configuration value for the"+
				"Graph API client id. Set the value statically in the configuration, or use the AZURE_CLIENT_ID environment variable.",
		)
	}

	if config.ClientSecret.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("client_secret"),
			"Unknown Graph client secret",
			"The provider cannot create the Graph API client secret as there is an unknown configuration value for the"+
				"Graph API client secret. Set the value statically in the configuration, or use the AZURE_CLIENT_SECRET environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but overridNewCdnDomainDataSourcee with Terraform
	// configuration value if set.
	var tenantID, clientID, clientSecret string

	if !config.TenantID.IsNull() {
		tenantID = config.TenantID.ValueString()
	} else {
		tenantID = os.Getenv("AZURE_TENANT_ID")
	}

	if !config.ClientID.IsNull() {
		clientID = config.ClientID.ValueString()
	} else {
		clientID = os.Getenv("AZURE_CLIENT_ID")
	}

	if !config.ClientSecret.IsNull() {
		clientSecret = config.ClientSecret.ValueString()
	} else {
		clientSecret = os.Getenv("AZURE_CLIENT_SECRET")
	}

	// If any of the expected configuration are missing, return errors with
	// provider-specific guidance.
	if tenantID == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("tenant_id"),
			"Missing Graph API tenant id",
			"The provider cannot create the Graph API client as there is a "+
				"missing or empty value for the Graph API tenant id. Set the "+
				"tenant value in the configuration or use the AZURE_TENANT_ID "+
				"environment variable. If either is already set, ensure the value "+
				"is not empty.",
		)
	}

	if clientID == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("client_id"),
			"Missing Graph API client id",
			"The provider cannot create the Graph API client as there is a "+
				"missing or empty value for the Graph API client id. Set the "+
				"client id value in the configuration or use the AZURE_CLIENT_ID "+
				"environment variable. If either is already set, ensure the value "+
				"is not empty.",
		)
	}

	if clientSecret == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("client_secret"),
			"Missing Graph API client secret",
			"The provider cannot create the Graph API client as there is a "+
				"missing or empty value for the Graph API client secret. Set the "+
				"client secret value in the configuration or use the AZURE_CLIENT_SECRET "+
				"environment variable. If either is already set, ensure the value "+
				"is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	cred, err := azureClient.NewClientSecretCredential(
		tenantID,
		clientID,
		clientSecret,
		nil,
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create MS Graph API Client",
			"An unexpected error occurred when creating the MS Graph API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"MS Graph API Client Error: "+err.Error(),
		)
		return
	}

	// MS Graph Client
	graphClient, err := graph.NewGraphServiceClientWithCredentials(
		cred, []string{"https://graph.microsoft.com/.default"},
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create MS Graph API Client",
			"An unexpected error occurred when creating the MS Graph API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"MS Graph API Client Error: "+err.Error(),
		)
		return
	}

	// Azuread clients wrapper
	azureadClients := azureadClients{
		graphClient: graphClient,
	}

	// Make the MS Graph API client available during DataSource and Resource type
	// Configure methods.
	resp.DataSourceData = azureadClients
	resp.ResourceData = azureadClients
}

func (p *azureadProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewAuthStrengthsDataSource,
		NewGroupsDataSource,
	}
}

func (p *azureadProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewNamedLocationResource,
	}
}
