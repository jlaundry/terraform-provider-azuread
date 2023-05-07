package applications

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/hashicorp/go-azure-sdk/sdk/odata"
	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-azuread/internal/clients"
	"github.com/hashicorp/terraform-provider-azuread/internal/tf"
	"github.com/hashicorp/terraform-provider-azuread/internal/utils"
	"github.com/hashicorp/terraform-provider-azuread/internal/validate"
	"github.com/manicminer/hamilton/msgraph"
)

func applicationIdentifierUrisResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: applicationIdentifierUrisResourceCreate,
		UpdateContext: applicationIdentifierUrisResourceUpdate,
		ReadContext:   applicationIdentifierUrisResourceRead,
		DeleteContext: applicationIdentifierUrisResourceDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
			Read:   schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Importer: tf.ValidateResourceIDPriorToImport(func(id string) error {
			if _, err := uuid.ParseUUID(id); err != nil {
				return fmt.Errorf("specified Application Object ID (%q) is not valid: %s", id, err)
			}
			return nil
		}),

		Schema: map[string]*schema.Schema{
			"application_object_id": {
				Description:      "The object ID of the application for which this federated identity credential should be created",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validate.UUID,
			},

			"identifier_uris": {
				Description: "The user-defined URI(s) that uniquely identify an application within its Azure AD tenant, or within a verified custom domain if the application is multi-tenant",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: validate.IsAppUri,
				},
			},
		},
	}
}

func applicationIdentifierUrisResourceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics { //nolint
	client := meta.(*clients.Client).Applications.ApplicationsClient
	objectId := d.Get("application_object_id").(string)

	tf.LockByName(applicationResourceName, objectId)
	defer tf.UnlockByName(applicationResourceName, objectId)

	app, status, err := client.Get(ctx, objectId, odata.Query{})
	if err != nil {
		if status == http.StatusNotFound {
			return tf.ErrorDiagPathF(nil, "application_object_id", "Application with object ID %q was not found", objectId)
		}
		return tf.ErrorDiagPathF(err, "application_object_id", "Retrieving application with object ID %q", objectId)
	}
	if app == nil || app.ID() == nil {
		return tf.ErrorDiagF(errors.New("nil application or application with nil ID was returned"), "API error retrieving application with object ID %q", objectId)
	}

	properties := msgraph.Application{
		DirectoryObject: msgraph.DirectoryObject{
			Id: utils.String(objectId),
		},
		IdentifierUris: tf.ExpandStringSlicePtr(d.Get("identifier_uris").(*schema.Set).List()),
	}

	if _, err := client.Update(ctx, properties); err != nil {
		return tf.ErrorDiagF(err, "Could not update application with object ID: %q", d.Id())
	}

	d.SetId(objectId)
	return applicationIdentifierUrisResourceRead(ctx, d, meta)
}

func applicationIdentifierUrisResourceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics { //nolint
	client := meta.(*clients.Client).Applications.ApplicationsClient

	tf.LockByName(applicationResourceName, d.Id())
	defer tf.UnlockByName(applicationResourceName, d.Id())

	app, status, err := client.Get(ctx, d.Id(), odata.Query{})
	if err != nil {
		if status == http.StatusNotFound {
			return tf.ErrorDiagPathF(nil, "application_object_id", "Application with object ID %q was not found", d.Id())
		}
		return tf.ErrorDiagPathF(err, "application_object_id", "Retrieving application with object ID %q", d.Id())
	}
	if app == nil || app.ID() == nil {
		return tf.ErrorDiagF(errors.New("nil application or application with nil ID was returned"), "API error retrieving application with object ID %q", d.Id())
	}

	properties := msgraph.Application{
		DirectoryObject: msgraph.DirectoryObject{
			Id: utils.String(d.Id()),
		},
		IdentifierUris: tf.ExpandStringSlicePtr(d.Get("identifier_uris").(*schema.Set).List()),
	}

	if _, err := client.Update(ctx, properties); err != nil {
		return tf.ErrorDiagF(err, "Could not update application with object ID: %q", d.Id())
	}

	return applicationIdentifierUrisResourceRead(ctx, d, meta)
}

func applicationIdentifierUrisResourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics { //nolint
	client := meta.(*clients.Client).Applications.ApplicationsClient

	app, status, err := client.Get(ctx, d.Id(), odata.Query{})
	if err != nil {
		if status == http.StatusNotFound {
			log.Printf("[DEBUG] Application with Object ID %q was not found - removing from state", d.Id())
			d.SetId("")
			return nil
		}

		return tf.ErrorDiagPathF(err, "id", "Retrieving Application with object ID %q", d.Id())
	}

	tf.Set(d, "identifier_uris", tf.FlattenStringSlicePtr(app.IdentifierUris))

	return nil
}

func applicationIdentifierUrisResourceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics { //nolint
	client := meta.(*clients.Client).Applications.ApplicationsClient

	tf.LockByName(applicationResourceName, d.Id())
	defer tf.UnlockByName(applicationResourceName, d.Id())

	app, status, err := client.Get(ctx, d.Id(), odata.Query{})
	if err != nil {
		if status == http.StatusNotFound {
			return tf.ErrorDiagPathF(nil, "application_object_id", "Application with object ID %q was not found", d.Id())
		}
		return tf.ErrorDiagPathF(err, "application_object_id", "Retrieving application with object ID %q", d.Id())
	}
	if app == nil || app.ID() == nil {
		return tf.ErrorDiagF(errors.New("nil application or application with nil ID was returned"), "API error retrieving application with object ID %q", d.Id())
	}

	properties := msgraph.Application{
		DirectoryObject: msgraph.DirectoryObject{
			Id: utils.String(d.Id()),
		},
		IdentifierUris: tf.ExpandStringSlicePtr(make([]interface{}, 0)),
	}

	if _, err := client.Update(ctx, properties); err != nil {
		return tf.ErrorDiagF(err, "Could not update application with object ID: %q", d.Id())
	}

	return nil
}
