package netlify

import (
	"errors"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/netlify/open-api/go/models"
	"github.com/netlify/open-api/go/plumbing/operations"
)

func resourceBranchDeploy() *schema.Resource {
	return &schema.Resource{
		Create: resourceBranchDeployCreate,
		Read:   resourceBranchDeployRead,
		Update: resourceBranchDeployUpdate,
		Delete: resourceBranchDeployDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"site_id": {
				Type:     schema.TypeString,
				Required: true,
			},

			"branch": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceBranchDeployCreate(d *schema.ResourceData, metaRaw interface{}) error {
	meta := metaRaw.(*Meta)
	b, branches, err := resourceBranchDeploy_getBranchAndBranches(d, meta)
	if err != nil {
		return err
	}

	branch := d.Get("branch").(string)

	for _, existing := range branches {
		if existing == branch {
			return errors.New(fmt.Sprintf("Branch deploy %s already exists", branch))
		}
	}
	branches = append(branches, branch, b)

	patch := operations.NewUpdateSiteParams()
	patch.SiteID = d.Get("site_id").(string)
	patch.Site = &models.SiteSetup{
		Site: models.Site{
			BuildSettings: &models.RepoInfo{
				AllowedBranches: branches,
			},
		},
	}
	_, err = meta.Netlify.Operations.UpdateSite(patch, meta.AuthInfo)
	d.SetId(branch)
	return err
}

func resourceBranchDeployRead(d *schema.ResourceData, metaRaw interface{}) error {
	meta := metaRaw.(*Meta)

	params := operations.NewGetSiteParams()
	params.SiteID = d.Get("site_id").(string)
	resp, err := meta.Netlify.Operations.GetSite(params, meta.AuthInfo)
	if err != nil {
		return err
	}

	for _, b := range resp.Payload.BuildSettings.AllowedBranches {
		if b == d.Get("branch").(string) {
			d.SetId(b)
			return nil
		}
	}

	d.SetId("")

	return nil
}

func resourceBranchDeployUpdate(d *schema.ResourceData, metaRaw interface{}) error {
	meta := metaRaw.(*Meta)

	oldBranch := d.Id()

	b, branches, err := resourceBranchDeploy_getBranchAndBranches(d, meta)
	if err != nil {
		return err
	}

	var newBranches []string
	for _, bb := range branches {
		if bb != oldBranch {
			newBranches = append(newBranches, bb)
		}
	}
	newBranches = append(newBranches, b, d.Get("branch").(string))

	params := operations.NewUpdateSiteParams()
	params.SiteID = d.Get("site_id").(string)

	params.Site = &models.SiteSetup{}
	params.Site.Repo = &models.RepoInfo{
		AllowedBranches: newBranches,
	}

	_, err = meta.Netlify.Operations.UpdateSite(params, meta.AuthInfo)

	return err
}

func resourceBranchDeployDelete(d *schema.ResourceData, metaRaw interface{}) error {
	meta := metaRaw.(*Meta)

	oldBranch := d.Id()

	b, branches, err := resourceBranchDeploy_getBranchAndBranches(d, meta)
	if err != nil {
		return err
	}

	var newBranches []string
	for _, bb := range branches {
		if bb != oldBranch {
			newBranches = append(newBranches, bb)
		}
	}
	newBranches = append(newBranches, b)

	params := operations.NewUpdateSiteParams()
	params.SiteID = d.Get("site_id").(string)

	params.Site = &models.SiteSetup{}
	params.Site.Repo = &models.RepoInfo{
		AllowedBranches: newBranches,
	}

	_, err = meta.Netlify.Operations.UpdateSite(params, meta.AuthInfo)

	return err
}

func resourceBranchDeploy_getBranchAndBranches(d *schema.ResourceData, meta *Meta) (string, []string, error) {
	params := operations.NewGetSiteParams()
	params.SiteID = d.Get("site_id").(string)
	resp, err := meta.Netlify.Operations.GetSite(params, meta.AuthInfo)
	if err != nil {
		return "", nil, err
	}
	branch := resp.Payload.BuildSettings.RepoBranch
	var branches []string
	for _, b := range resp.Payload.BuildSettings.AllowedBranches {
		if b != branch {
			branches = append(branches, branch)
		}
	}
	return branch, branches, nil
}
