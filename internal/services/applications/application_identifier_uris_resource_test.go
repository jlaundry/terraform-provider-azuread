package applications_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/go-azure-sdk/sdk/odata"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-azuread/internal/acceptance"
	"github.com/hashicorp/terraform-provider-azuread/internal/acceptance/check"
	"github.com/hashicorp/terraform-provider-azuread/internal/clients"
	"github.com/hashicorp/terraform-provider-azuread/internal/services/applications/parse"
	"github.com/hashicorp/terraform-provider-azuread/internal/utils"
)

type ApplicationIdentifierUrisResource struct{}

func TestAccApplicationIdentifierUris_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azuread_application_identifier_uris", "test")
	r := ApplicationIdentifierUrisResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.basic(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("identifier_uris").Exists(),
			),
		},
		data.ImportStep(),
	})
}

func TestAccApplicationIdentifierUris_complete(t *testing.T) {
	data := acceptance.BuildTestData(t, "azuread_application_identifier_uris", "test")
	r := ApplicationIdentifierUrisResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.complete(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("identifier_uris").Exists(),
			),
		},
		data.ImportStep(),
	})
}

func TestAccApplicationIdentifierUris_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "azuread_application_identifier_uris", "test")
	r := ApplicationIdentifierUrisResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.basic(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("identifier_uris").Exists(),
			),
		},
		data.ImportStep(),
		{
			Config: r.complete(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("identifier_uris").Exists(),
			),
		},
		data.ImportStep(),
		{
			Config: r.basic(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("identifier_uris").Exists(),
			),
		},
		data.ImportStep(),
	})
}

func (r ApplicationIdentifierUrisResource) Exists(ctx context.Context, clients *clients.Client, state *terraform.InstanceState) (*bool, error) {
	client := clients.Applications.ApplicationsClient
	client.BaseClient.DisableRetries = true
	defer func() { client.BaseClient.DisableRetries = false }()

	id, err := parse.FederatedIdentityCredentialID(state.ID)
	if err != nil {
		return nil, fmt.Errorf("parsing Application Identifier URIs ID: %v", err)
	}

	credential, status, err := client.GetFederatedIdentityCredential(ctx, id.ObjectId, id.KeyId, odata.Query{})
	if err != nil {
		if status == http.StatusNotFound {
			return nil, fmt.Errorf("Identifier URIs %q for Application with object ID %q does not exist", id.KeyId, id.ObjectId)
		}
		return nil, fmt.Errorf("failed to retrieve Identifier URIs %q for Application with object ID %q: %+v", id.KeyId, id.ObjectId, err)
	}

	return utils.Bool(credential != nil), nil
}

func (ApplicationIdentifierUrisResource) template(data acceptance.TestData) string {
	return fmt.Sprintf(`
resource "azuread_application" "test" {
  display_name = "acctestApplicationIdentifierUris-%[1]d"
}
`, data.RandomInteger)
}

func (r ApplicationIdentifierUrisResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azuread_application_identifier_uris" "test" {
  application_object_id = azuread_application.test.object_id
  identifier_uris       = ["api://${azuread_application.test.id}"]
}
`, r.template(data), data.RandomString, data.RandomID)
}

func (r ApplicationIdentifierUrisResource) complete(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

data "azuread_domains" "aad_domains" {
  only_default = true
}

resource "azuread_application_identifier_uris" "test" {
  application_object_id = azuread_application.test.object_id
  identifier_uris       = ["api://${azuread_application.test.id}", "https://example.${data.azuread_domains.aad_domains.domains.0.domain_name}"]
}
`, r.template(data), data.RandomString, data.RandomID)
}
