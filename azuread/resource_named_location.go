package azuread

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"

	graph "github.com/microsoftgraph/msgraph-sdk-go"
	graphmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
)

const (
	IPOdataType      = "#microsoft.graph.ipNamedLocation"
	CountryOdataType = "#microsoft.graph.countryNamedLocation"
	IPV4Type         = "#microsoft.graph.iPv4CidrRange"
	IPV6Type         = "#microsoft.graph.iPv6CidrRange"
	IDPath           = "/identity/conditionalAccess/namedLocations/"
)

var (
	_ resource.Resource              = &namedLocationResource{}
	_ resource.ResourceWithConfigure = &namedLocationResource{}
)

func NewNamedLocationResource() resource.Resource {
	return &namedLocationResource{}
}

type namedLocationResource struct {
	client *graph.GraphServiceClient
}

type namedLocationResourceModel struct {
	ID          types.String  `tfsdk:"id"`
	DisplayName types.String  `tfsdk:"display_name"`
	IP          *ipBlock      `tfsdk:"ip"`
	Country     *countryBlock `tfsdk:"country"`
}

type ipBlock struct {
	IPRanges     types.List `tfsdk:"ip_ranges"`
	Trusted      types.Bool `tfsdk:"trusted"`
	ForceDestroy types.Bool `tfsdk:"force_destroy"`
}

type countryBlock struct {
	CountriesAndRegions               types.List   `tfsdk:"countries_and_regions"`
	CountryLookupMethod               types.String `tfsdk:"country_lookup_method"`
	IncludeUnknownCountriesAndRegions types.Bool   `tfsdk:"include_unknown_countries_and_regions"`
}

func (r *namedLocationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_named_location"
}

func (r *namedLocationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Named Location within Azure Active Directory.",
		Attributes: map[string]schema.Attribute{
			"display_name": schema.StringAttribute{
				Required:    true,
				Description: "The display name for the named location.",
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID for the named location.",
			},
		},
		Blocks: map[string]schema.Block{
			"ip": schema.SingleNestedBlock{
				Description: "IP-based named location settings.",
				Attributes: map[string]schema.Attribute{
					"ip_ranges": schema.ListAttribute{
						ElementType: types.StringType,
						Optional:    true,
						Description: "List of IP ranges in CIDR format.",
					},
					"trusted": schema.BoolAttribute{
						Optional:    true,
						Description: "Whether the named location is trusted. Defaults to 'false'.",
					},
					"force_destroy": schema.BoolAttribute{
						Optional:    true,
						Description: "Whether to forcefully destroy the IP-based named location when trusted is set to true. This prevents accidental deletion of trusted IPs.",
					},
				},
			},
			"country": schema.SingleNestedBlock{
				Description: "Country-based named location settings.",
				Attributes: map[string]schema.Attribute{
					"countries_and_regions": schema.ListAttribute{
						ElementType: types.StringType,
						Optional:    true,
						Description: "List of countries and/or regions in two-letter format specified by ISO 3166-2.",
						Validators: []validator.List{
							listvalidator.ValueStringsAre(stringvalidator.OneOf("ZW", "ZM", "YE", "EH", "WF", "VI", "VG", "VN", "VE", "VU", "UZ", "UM", "UY", "US", "GB",
								"AE", "UA", "UG", "TV", "TC", "TM", "TR", "TN", "TT", "TO", "TK", "TG", "TL", "TH", "TZ",
								"TJ", "TW", "SY", "CH", "SE", "SJ", "SR", "SD", "LK", "ES", "SS", "GS", "ZA", "SO", "SB",
								"SI", "SK", "SX", "SG", "SL", "SC", "RS", "SN", "SA", "ST", "SM", "WS", "VC", "PM", "MF",
								"LC", "KN", "SH", "BL", "RW", "RU", "RO", "RE", "CG", "QA", "PR", "PT", "PL", "PN", "PH",
								"PE", "PY", "PG", "PA", "PS", "PW", "PK", "OM", "NO", "MP", "MK", "KP", "NF", "NU", "NG",
								"NE", "NI", "NZ", "NC", "NL", "NP", "NR", "NA", "MM", "MZ", "MA", "MS", "ME", "MN", "MC",
								"MD", "FM", "MX", "YT", "MU", "MR", "MQ", "MH", "MT", "ML", "MV", "MY", "MW", "MG", "MO",
								"LU", "LT", "LI", "LY", "LR", "LS", "LB", "LV", "LA", "KG", "KW", "XK", "KR", "KI", "KE",
								"KZ", "JO", "JE", "JP", "JM", "IT", "IL", "IM", "IE", "IQ", "IR", "ID", "IN", "IS", "HU",
								"HK", "HN", "VA", "HM", "HT", "GY", "GW", "GN", "GG", "GT", "GU", "GP", "GD", "GL", "GR",
								"GI", "GH", "DE", "GE", "GM", "GA", "TF", "PF", "GF", "FR", "FI", "FJ", "FO", "FK", "ET",
								"SZ", "EE", "ER", "GQ", "SV", "EG", "EC", "DO", "DM", "DJ", "DK", "CD", "CZ", "CY", "CW",
								"CU", "HR", "CI", "CR", "CK", "KM", "CO", "CC", "CX", "CN", "CL", "TD", "CF", "KY", "CA",
								"CM", "KH", "CV", "BI", "BF", "BG", "BN", "IO", "BR", "BV", "BW", "BA", "BQ", "BO", "BT",
								"BM", "BJ", "BZ", "BE", "BY", "BB", "BD", "BH", "BS", "AZ", "AT", "AU", "AW", "AM", "AR",
								"AG", "AQ", "AI", "AO", "AD", "AS", "DZ", "AL", "AX", "AF")),
						},
					},
					"include_unknown_countries_and_regions": schema.BoolAttribute{
						Optional:    true,
						Description: "Whether IP addresses that don't map to a country or region should be included in the named location. Defaults to 'false'.",
					},
					"country_lookup_method": schema.StringAttribute{
						Optional:    true,
						Description: "Method of detecting country the user is located in. Possible values are 'clientIpAddress' for IP-based location and 'authenticatorAppGps' for Authenticator app GPS-based location. Defaults to 'clientIpAddress'.",
						Validators: []validator.String{
							stringvalidator.OneOf("clientIpAddress", "authenticatorAppGps"),
						},
					},
				},
			},
		},
	}
}

func (r *namedLocationResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(azureadClients).graphClient
}

func (r *namedLocationResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// If the entire plan is null, the resource is planned for destruction.
	if !req.Plan.Raw.IsNull() {
		var plan namedLocationResourceModel
		diags := req.Plan.Get(ctx, &plan)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		// Validate that if ip.trusted is true, ip.force_destroy must be explicitly set.
		if plan.IP != nil &&
			!plan.IP.Trusted.IsNull() &&
			!plan.IP.Trusted.IsUnknown() &&
			plan.IP.Trusted.ValueBool() {

			if plan.IP.ForceDestroy.IsNull() || plan.IP.ForceDestroy.IsUnknown() {
				resp.Diagnostics.AddAttributeError(
					path.Root("ip").AtName("force_destroy"),
					"[INPUT ERROR] Missing Input",
					"When 'ip.trusted' is set to true, you must also set 'ip.force_destroy' explicitly.",
				)
				return
			}
		}
	}
}

func (r *namedLocationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan namedLocationResourceModel
	diags := req.Config.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// To check whether is country-based named location or IP-based named location.
	if plan.IP != nil {
		err := r.createIPNamedLocation(&plan)
		if err != nil {
			resp.Diagnostics.AddError(
				"[API ERROR] Failed to create IP named location.",
				err.Error(),
			)
			return
		}
	} else if plan.Country != nil {
		err := r.createCountryNamedLocation(&plan, resp)
		if err != nil {
			resp.Diagnostics.AddError(
				"[API ERROR] Failed to create country named location.",
				err.Error(),
			)
			return
		}
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *namedLocationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state namedLocationResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Since the ID in state is '/identity/conditionalAccess/namedLocations/{id}', the ID use in here need to be ID only whitout '/identity/conditionalAccess/namedLocations/'.
	fullID := state.ID.ValueString()
	parts := strings.Split(fullID, "/")
	namedLocationId := parts[len(parts)-1]

	var result graphmodels.NamedLocationable
	operation := func() error {
		var err error
		result, err = r.client.Identity().
			ConditionalAccess().
			NamedLocations().
			ByNamedLocationId(namedLocationId).
			Get(ctx, nil)
		if err != nil {
			// You can add logic here to determine if the error is retryable
			return err
		}
		if result == nil || result.GetOdataType() == nil {
			return fmt.Errorf("Named location does not exist or has an invalid @odata.type.")
		}
		return nil
	}

	// Configure exponential backoff parameters
	retryBackoff := backoff.NewExponentialBackOff()
	retryBackoff.MaxElapsedTime = 30 * time.Second

	err := backoff.Retry(operation, backoff.WithContext(retryBackoff, ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Microsoft Graph Error",
			fmt.Sprintf("Failed to retrieve named location: %s", err.Error()),
		)
		return
	}

	odataType := *result.GetOdataType()
	switch odataType {
	case CountryOdataType:
		countryNamedLocation, ok := result.(graphmodels.CountryNamedLocationable)
		if !ok {
			resp.Diagnostics.AddError(
				"Type Assertion Failed",
				"Could not convert result to CountryNamedLocationable.",
			)
			return
		}

		if val := countryNamedLocation.GetDisplayName(); val != nil {
			state.DisplayName = types.StringValue(*val)
		}
		if val := countryNamedLocation.GetIncludeUnknownCountriesAndRegions(); val != nil {
			state.Country.IncludeUnknownCountriesAndRegions = types.BoolValue(*val)
		}
		if val := countryNamedLocation.GetCountryLookupMethod(); val != nil {
			var method string

			switch *val {
			case graphmodels.CountryLookupMethodType(0):
				method = "clientIpAddress"
			case graphmodels.CountryLookupMethodType(1):
				method = "authenticatorAppGps"
			case graphmodels.CountryLookupMethodType(2):
				method = "unknownFutureValue"
			default:
				method = "unknown"
			}

			state.Country.CountryLookupMethod = types.StringValue(method)
		}

		countryList := countryNamedLocation.GetCountriesAndRegions()
		var countriesAndRegions []types.String
		for _, c := range countryList {
			countriesAndRegions = append(countriesAndRegions, types.StringValue(c))
		}
		state.Country.CountriesAndRegions, diags = types.ListValueFrom(ctx, types.StringType, countriesAndRegions)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		state.IP = nil

	case IPOdataType:
		ipNamedLocation, ok := result.(graphmodels.IpNamedLocationable)
		if !ok {
			resp.Diagnostics.AddError(
				"Type Assertion Failed",
				"Could not convert result to IpNamedLocationable.",
			)
			return
		}

		if val := ipNamedLocation.GetDisplayName(); val != nil {
			state.DisplayName = types.StringValue(*val)
		}
		if val := ipNamedLocation.GetIsTrusted(); val != nil {
			state.IP.Trusted = types.BoolValue(*val)
		}

		ipList := ipNamedLocation.GetIpRanges()
		var ipStrings []types.String

		for _, ip := range ipList {
			switch v := ip.(type) {
			case graphmodels.IPv4CidrRangeable:
				if cidr := v.GetCidrAddress(); cidr != nil {
					ipStrings = append(ipStrings, types.StringValue(*cidr))
				}
			case graphmodels.IPv6CidrRangeable:
				if cidr := v.GetCidrAddress(); cidr != nil {
					ipStrings = append(ipStrings, types.StringValue(*cidr))
				}
			default:
				resp.Diagnostics.AddError(
					"Unsupported IP Range Type",
					fmt.Sprintf("Unknown IP range type: %T", v),
				)
				return
			}
		}

		state.IP.IPRanges, diags = types.ListValueFrom(ctx, types.StringType, ipStrings)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		state.Country = nil

	default:
		resp.Diagnostics.AddError(
			"Unsupported Named Location Type",
			fmt.Sprintf("Unexpected @odata.type '%s' for named location.", odataType),
		)
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *namedLocationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan namedLocationResourceModel
	getPlanDiags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(getPlanDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state namedLocationResourceModel
	getStateDiags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(getStateDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.IP != nil {
		state.IP.ForceDestroy = plan.IP.ForceDestroy
		if err := r.updateIpNamedLocation(ctx, &plan, &state); err != nil {
			resp.Diagnostics.AddError(
				"[API ERROR] Failed to update named location.",
				err.Error(),
			)
			return
		}
	} else if plan.Country != nil {
		if err := r.updateCountryNamedLocation(ctx, &plan, &state, (*resource.CreateResponse)(resp)); err != nil {
			resp.Diagnostics.AddError(
				"[API ERROR] Failed to update named location.",
				err.Error(),
			)
			return
		}
	}

	plan.ID = state.ID
	resp.State.Set(ctx, &plan)
}

func (r *namedLocationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state namedLocationResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	fullID := state.ID.ValueString()
	parts := strings.Split(fullID, "/")
	namedLocationID := parts[len(parts)-1]

	if state.IP != nil && state.IP.Trusted.ValueBool() && !state.IP.ForceDestroy.ValueBool() {
		resp.Diagnostics.AddError(
			"Unable to delete Named Location",
			fmt.Sprintf(
				"Named Location %q (ID: %q) cannot be deleted because it is marked as a Trusted location. "+
					"To delete this resource, you must explicitly set the 'ip.force_destroy' attribute to true.",
				state.DisplayName.ValueString(),
				namedLocationID,
			),
		)
		return
	}

	// If it's an IP-based named location and it's trusted and the force destroy is set to true, set IsTrusted to false before deletion.
	if state.IP != nil && state.IP.Trusted.ValueBool() && state.IP.ForceDestroy.ValueBool() {
		isTrusted := false
		updateIpNamedLocationRequest := graphmodels.NewIpNamedLocation()
		updateIpNamedLocationRequest.SetIsTrusted(&isTrusted)
		displayName := state.DisplayName.ValueString()
		updateIpNamedLocationRequest.SetDisplayName(&displayName)

		var ipRanges []graphmodels.IpRangeable

		for _, ipVal := range state.IP.IPRanges.Elements() {
			cidrStrVal, ok := ipVal.(types.String)
			if !ok {
				resp.Diagnostics.AddError(
					"Invalid IP Range Type",
					"Expected a string for IP range, but got a different type",
				)
				return
			}

			cidr := cidrStrVal.ValueString()
			ip, _, err := net.ParseCIDR(cidr)
			if err != nil {
				resp.Diagnostics.AddError(
					"Invalid CIDR Format",
					fmt.Sprintf("Failed to parse CIDR: %s", cidr),
				)
				return
			}

			if ip.To4() != nil {
				ipv4Range := graphmodels.NewIPv4CidrRange()
				ipv4Range.SetCidrAddress(&cidr)
				ipv4Type := IPV4Type
				ipv4Range.SetOdataType(&ipv4Type)
				ipRanges = append(ipRanges, ipv4Range)
			} else {
				ipv6Range := graphmodels.NewIPv6CidrRange()
				ipv6Range.SetCidrAddress(&cidr)
				ipv6Type := IPV6Type
				ipv6Range.SetOdataType(&ipv6Type)
				ipRanges = append(ipRanges, ipv6Range)
			}
		}

		updateIpNamedLocationRequest.SetIpRanges(ipRanges)
		odataType := IPOdataType
		updateIpNamedLocationRequest.SetOdataType(&odataType)

		// Backoff retry for the PATCH request.
		var patchErr error
		backoff := 2 * time.Second
		for i := 0; i < 5; i++ {
			_, patchErr = r.client.Identity().
				ConditionalAccess().
				NamedLocations().
				ByNamedLocationId(namedLocationID).
				Patch(ctx, updateIpNamedLocationRequest, nil)

			if patchErr == nil {
				break
			}

			time.Sleep(backoff)
			backoff *= 2
		}

		if patchErr != nil {
			resp.Diagnostics.AddError(
				"Unable to update Named Location",
				fmt.Sprintf("Failed to update Named Location with ID %q after %d retries: %s", namedLocationID, 5, patchErr.Error()),
			)
			return
		}

		time.Sleep(1 * time.Minute)
	}

	err := r.client.Identity().
		ConditionalAccess().
		NamedLocations().
		ByNamedLocationId(namedLocationID).
		Delete(ctx, nil)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete Named Location",
			fmt.Sprintf("Error deleting Named Location with ID %q: %s", namedLocationID, err.Error()),
		)
		return
	}

	time.Sleep(1 * time.Minute)
}

func (d *namedLocationResource) createIPNamedLocation(plan *namedLocationResourceModel) error {
	createIPNamedLocationRequest := graphmodels.NewIpNamedLocation()
	odataType := IPOdataType
	createIPNamedLocationRequest.SetOdataType(&odataType)
	createIPNamedLocationRequest.SetDisplayName(plan.DisplayName.ValueStringPointer())
	if plan.IP.Trusted.IsNull() || plan.IP.Trusted.IsUnknown() {
		return fmt.Errorf("The field 'Trusted' must be set and cannot be null or unknown.")
	} else {
		createIPNamedLocationRequest.SetIsTrusted(plan.IP.Trusted.ValueBoolPointer())
	}

	var ipRanges []graphmodels.IpRangeable

	for _, attrVal := range plan.IP.IPRanges.Elements() {
		cidrStrVal, ok := attrVal.(types.String)
		if !ok {
			return fmt.Errorf("invalid IP range type, expected string")
		}

		cidr := cidrStrVal.ValueString()

		ip, _, err := net.ParseCIDR(cidr)
		if err != nil {
			return fmt.Errorf("invalid CIDR format: %s", cidr)
		}

		if ip.To4() != nil {
			ipv4Range := graphmodels.NewIPv4CidrRange()
			ipv4Range.SetCidrAddress(&cidr)

			ipv4Type := IPV4Type
			ipv4Range.SetOdataType(&ipv4Type)

			ipRanges = append(ipRanges, ipv4Range)
		} else {
			ipv6Range := graphmodels.NewIPv6CidrRange()
			ipv6Range.SetCidrAddress(&cidr)

			ipv6Type := IPV6Type
			ipv6Range.SetOdataType(&ipv6Type)

			ipRanges = append(ipRanges, ipv6Range)
		}
	}

	createIPNamedLocationRequest.SetIpRanges(ipRanges)

	createIPNamedLocation := func() error {
		created, err := d.client.Identity().
			ConditionalAccess().
			NamedLocations().
			Post(
				context.Background(),
				createIPNamedLocationRequest,
				nil,
			)
		if err != nil {
			if t, ok := err.(*errors.TencentCloudSDKError); ok {
				codeStr := t.GetCode()
				codeInt, convErr := strconv.Atoi(codeStr)
				if convErr != nil {
					return backoff.Permanent(fmt.Errorf("failed to convert error code to int: %v", convErr))
				}
				if isAbleToRetry(codeInt) {
					return err // retryable
				}
				return backoff.Permanent(err) // permanent
			}
			return err
		}

		if created.GetId() == nil || *created.GetId() == "" {
			return backoff.Permanent(fmt.Errorf("created IPNamedLocation returned no ID"))
		}

		plan.ID = types.StringValue(IDPath + *created.GetId())
		time.Sleep(1 * time.Minute) // Azure propagation delay
		return nil
	}

	// First attempt without retry
	if err := createIPNamedLocation(); err == nil {
		return nil
	}

	// Retry with exponential backoff
	retryBackoff := backoff.NewExponentialBackOff()
	retryBackoff.MaxElapsedTime = 30 * time.Second

	return backoff.Retry(func() error {
		// Calls the Microsoft Graph SDK to get a list of existing conditional access named locations.
		existingLocations, err := d.client.Identity().
			ConditionalAccess().
			NamedLocations().
			Get(context.Background(), nil)
		if err != nil {
			return err
		}

		for _, location := range existingLocations.GetValue() {
			if location.GetId() != nil && plan.ID.ValueString() == IDPath+*location.GetId() {
				return nil // Found the location with matching ID; stop retrying.
			}
		}

		// If no matching location is found, Calls createIPNamedLocation() to attempt to create the named location.
		return createIPNamedLocation()
	}, retryBackoff)
}

func (d *namedLocationResource) createCountryNamedLocation(plan *namedLocationResourceModel, resp *resource.CreateResponse) error {
	createCountryNamedLocationRequest := graphmodels.NewCountryNamedLocation()
	odataType := CountryOdataType
	createCountryNamedLocationRequest.SetOdataType(&odataType)
	createCountryNamedLocationRequest.SetDisplayName(plan.DisplayName.ValueStringPointer())

	if plan.Country.IncludeUnknownCountriesAndRegions.IsNull() || plan.Country.IncludeUnknownCountriesAndRegions.IsUnknown() {
		return fmt.Errorf("The field 'IncludeUnknownCountriesAndRegions' must be set and cannot be null or unknown.")
	} else {
		createCountryNamedLocationRequest.SetIncludeUnknownCountriesAndRegions(plan.Country.IncludeUnknownCountriesAndRegions.ValueBoolPointer())
	}

	var countriesAndRegions []string

	for _, countryAndRegion := range plan.Country.CountriesAndRegions.Elements() {
		strVal, ok := countryAndRegion.(types.String)
		if !ok {
			return fmt.Errorf("expected types.String but got %T", countryAndRegion)
		}
		countriesAndRegions = append(countriesAndRegions, strVal.ValueString())
	}

	createCountryNamedLocationRequest.SetCountriesAndRegions(countriesAndRegions)

	var diags diag.Diagnostics
	var method graphmodels.CountryLookupMethodType
	lookupMethod := plan.Country.CountryLookupMethod

	if lookupMethod.IsNull() || lookupMethod.IsUnknown() {
		method = graphmodels.CountryLookupMethodType(0)
	} else {
		switch lookupMethod.ValueString() {
		case "clientIpAddress":
			method = graphmodels.CountryLookupMethodType(0)
		case "authenticatorAppGps":
			method = graphmodels.CountryLookupMethodType(1)
		case "unknownFutureValue":
			method = graphmodels.CountryLookupMethodType(2)
		default:
			diags.AddError(
				"Invalid Country Lookup Method",
				fmt.Sprintf("Unknown lookup method: %s", lookupMethod.ValueString()),
			)
			resp.Diagnostics.Append(diags...)
			return fmt.Errorf("invalid country lookup method: %s", lookupMethod.ValueString())
		}
	}

	createCountryNamedLocationRequest.SetCountryLookupMethod(&method)

	createCountryNamedLocation := func() error {
		created, err := d.client.Identity().
			ConditionalAccess().
			NamedLocations().
			Post(
				context.Background(),
				createCountryNamedLocationRequest,
				nil,
			)
		if err != nil {
			if t, ok := err.(*errors.TencentCloudSDKError); ok {
				codeStr := t.GetCode()
				codeInt, convErr := strconv.Atoi(codeStr)
				if convErr != nil {
					return backoff.Permanent(fmt.Errorf("failed to convert error code to int: %v", convErr))
				}
				if isAbleToRetry(codeInt) {
					return err // retryable
				}
				return backoff.Permanent(err) // permanent
			}
			return err
		}

		if created.GetId() == nil || *created.GetId() == "" {
			return backoff.Permanent(fmt.Errorf("created CountryNamedLocation returned no ID"))
		}

		plan.ID = types.StringValue(IDPath + *created.GetId())
		time.Sleep(1 * time.Minute) // Azure propagation delay
		return nil
	}

	// First attempt without retry
	if err := createCountryNamedLocation(); err == nil {
		return nil
	}

	// Retry with exponential backoff
	retryBackoff := backoff.NewExponentialBackOff()
	retryBackoff.MaxElapsedTime = 30 * time.Second

	return backoff.Retry(func() error {
		existingLocations, err := d.client.Identity().
			ConditionalAccess().
			NamedLocations().
			Get(context.Background(), nil)
		if err != nil {
			return err
		}

		for _, location := range existingLocations.GetValue() {
			if location.GetId() != nil && plan.ID.ValueString() == IDPath+*location.GetId() {
				return nil // ID found, no need to retry
			}
		}

		return createCountryNamedLocation()
	}, retryBackoff)
}

func (d *namedLocationResource) updateIpNamedLocation(ctx context.Context, plan, state *namedLocationResourceModel) error {
	updateIpNamedLocationRequest := graphmodels.NewIpNamedLocation()

	// Extract the ID from previous state.
	fullID := state.ID.ValueString()
	parts := strings.Split(fullID, "/")
	namedLocationId := parts[len(parts)-1]

	odataType := IPOdataType
	updateIpNamedLocationRequest.SetOdataType(&odataType)

	displayName := plan.DisplayName.ValueString()
	updateIpNamedLocationRequest.SetDisplayName(&displayName)

	updateIpNamedLocationRequest.SetIsTrusted(
		plan.IP.Trusted.ValueBoolPointer(),
	)

	var ipRanges []graphmodels.IpRangeable

	for _, ipVal := range plan.IP.IPRanges.Elements() {
		cidrStrVal, ok := ipVal.(types.String)
		if !ok {
			return fmt.Errorf("invalid IP range type, expected string")
		}

		cidr := cidrStrVal.ValueString()

		// Parse the CIDR to determine whether it's IPv4 or IPv6.
		ip, _, err := net.ParseCIDR(cidr)
		if err != nil {
			return fmt.Errorf("invalid CIDR format: %s", cidr)
		}

		if ip.To4() != nil {
			ipv4Range := graphmodels.NewIPv4CidrRange()
			ipv4Range.SetCidrAddress(&cidr)

			ipv4Type := IPV4Type
			ipv4Range.SetOdataType(&ipv4Type)

			ipRanges = append(ipRanges, ipv4Range)
		} else {
			ipv6Range := graphmodels.NewIPv6CidrRange()
			ipv6Range.SetCidrAddress(&cidr)

			ipv6Type := IPV6Type
			ipv6Range.SetOdataType(&ipv6Type)

			ipRanges = append(ipRanges, ipv6Range)
		}
	}

	updateIpNamedLocationRequest.SetIpRanges(ipRanges)

	updateIpNamedLocation := func() error {
		_, err := d.client.Identity().
			ConditionalAccess().
			NamedLocations().
			ByNamedLocationId(namedLocationId).
			Patch(ctx, updateIpNamedLocationRequest, nil)

		if err != nil {
			return err
		}
		return nil
	}

	reconnectBackoff := backoff.NewExponentialBackOff()
	reconnectBackoff.MaxElapsedTime = 30 * time.Second
	return backoff.Retry(updateIpNamedLocation, reconnectBackoff)
}

func (d *namedLocationResource) updateCountryNamedLocation(ctx context.Context, plan, state *namedLocationResourceModel, resp *resource.CreateResponse) error {
	updateCountryNamedLocationRequest := graphmodels.NewCountryNamedLocation()

	// Extract the ID from previous state.
	fullID := state.ID.ValueString()
	parts := strings.Split(fullID, "/")
	namedLocationId := parts[len(parts)-1]

	odataType := CountryOdataType
	updateCountryNamedLocationRequest.SetOdataType(&odataType)

	displayName := plan.DisplayName.ValueString()
	updateCountryNamedLocationRequest.SetDisplayName(&displayName)

	trusted := plan.Country.IncludeUnknownCountriesAndRegions.ValueBoolPointer()
	updateCountryNamedLocationRequest.SetIncludeUnknownCountriesAndRegions(trusted)

	var diags diag.Diagnostics
	var method graphmodels.CountryLookupMethodType
	lookupMethod := plan.Country.CountryLookupMethod

	if lookupMethod.IsNull() || lookupMethod.IsUnknown() {
		method = graphmodels.CountryLookupMethodType(0)
	} else {
		switch lookupMethod.ValueString() {
		case "clientIpAddress":
			method = graphmodels.CountryLookupMethodType(0)
		case "authenticatorAppGps":
			method = graphmodels.CountryLookupMethodType(1)
		case "unknownFutureValue":
			method = graphmodels.CountryLookupMethodType(2)
		default:
			diags.AddError(
				"Invalid Country Lookup Method",
				fmt.Sprintf("Unknown lookup method: %s", lookupMethod.ValueString()),
			)
			resp.Diagnostics.Append(diags...)
			return fmt.Errorf("invalid country lookup method: %s", lookupMethod.ValueString())
		}
	}

	updateCountryNamedLocationRequest.SetCountryLookupMethod(&method)

	var countries []string
	for _, countryVal := range plan.Country.CountriesAndRegions.Elements() {
		strVal, ok := countryVal.(types.String)
		if !ok {
			return fmt.Errorf("expected types.String but got %T", countryVal)
		}
		countries = append(countries, strVal.ValueString())
	}
	updateCountryNamedLocationRequest.SetCountriesAndRegions(countries)

	updateCountryNamedLocation := func() error {
		_, err := d.client.Identity().
			ConditionalAccess().
			NamedLocations().
			ByNamedLocationId(namedLocationId).
			Patch(ctx, updateCountryNamedLocationRequest, nil)

		if err != nil {
			return err
		}
		return nil
	}

	reconnectBackoff := backoff.NewExponentialBackOff()
	reconnectBackoff.MaxElapsedTime = 30 * time.Second
	return backoff.Retry(updateCountryNamedLocation, reconnectBackoff)
}
