package azuread

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/cenkalti/backoff"
	graph "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &authStrengthsDataSource{}
	_ datasource.DataSourceWithConfigure = &authStrengthsDataSource{}
)

func NewAuthStrengthsDataSource() datasource.DataSource {
	return &authStrengthsDataSource{}
}

type authStrengthsDataSource struct {
	client *graph.GraphServiceClient
}

type authStrengthsDataSourceModel struct {
	AuthStrIDs    types.List           `tfsdk:"ids"`
	AuthStrNames  types.List           `tfsdk:"names"`
	AuthStrengths []*authStrengthModel `tfsdk:"auth_strengths"`
}

type authStrengthModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func (d *authStrengthsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_auth_strengths"
}

func (d *authStrengthsDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This data source provides the authentication strength policies based on the list of ids" +
			" or names of the policies. Will return all policies if input is empty.",
		Attributes: map[string]schema.Attribute{
			"ids": schema.ListAttribute{
				Description: "The IDs of the authentication strength policy.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"names": schema.ListAttribute{
				Description: "The names of the authentication strength policy.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"auth_strengths": schema.ListNestedAttribute{
				Description: "List of authentication strength policies with ID and name.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "ID of the authentication strength policy.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Display name of the authentication strength policy.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *authStrengthsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(azureadClients).graphClient
}

func (d *authStrengthsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var authStrengths models.AuthenticationStrengthPolicyCollectionResponseable
	var err error
	var plan authStrengthsDataSourceModel

	diags := req.Config.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if len(plan.AuthStrIDs.Elements()) > 0 && len(plan.AuthStrNames.Elements()) > 0 {
		resp.Diagnostics.AddError(
			"[INPUT ERROR] Invalid Input",
			"Only one of 'ids' or 'names' may be specified, not both.",
		)
		return
	}

	// Retrieve authentication strength policies with backoff retry.
	getAuthStrengths := func() error {
		authStrengths, err = d.client.Policies().AuthenticationStrengthPolicies().Get(context.Background(), nil)
		return handleAPIError(err)
	}

	backoffPolicy := backoff.NewExponentialBackOff()
	backoffPolicy.MaxElapsedTime = 30 * time.Second
	if err = backoff.Retry(getAuthStrengths, backoffPolicy); err != nil {
		resp.Diagnostics.AddError(
			"[API ERROR] Unable to Retrieve Authentication Strength Policy",
			"An error occurred while retrieving Authentication Strength Policies from Microsoft Graph API. "+
				"Check the policy names and API permissions.\n\nGraph API Error: "+err.Error(),
		)
		return
	}

	nameSlice := make([]string, len(plan.AuthStrNames.Elements()))
	for i, o := range plan.AuthStrNames.Elements() {
		nameSlice[i] = strings.Trim(o.(types.String).ValueString(), "\"")
	}

	idSlice := make([]string, len(plan.AuthStrIDs.Elements()))
	for i, o := range plan.AuthStrIDs.Elements() {
		idSlice[i] = strings.Trim(o.(types.String).ValueString(), "\"")
	}

	var authStrengthItems []*authStrengthModel
	for _, policy := range authStrengths.GetValue() {
		namePtr := policy.GetDisplayName()
		idPtr := policy.GetId()
		name := *namePtr
		id := *idPtr
		fullID := fmt.Sprintf("/policies/authenticationStrengthPolicies/%s", id)

		if (len(nameSlice) > 0 && !slices.Contains(nameSlice, name)) ||
			(len(idSlice) > 0 && !slices.Contains(idSlice, id)) {
			continue
		}

		authStrengthItems = append(authStrengthItems, &authStrengthModel{
			ID:   types.StringValue(fullID),
			Name: types.StringValue(name),
		})
	}

	plan.AuthStrengths = authStrengthItems
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}
