package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		// Get projects collection for relation
		projects, err := app.FindCollectionByNameOrId("projects")
		if err != nil {
			return err
		}

		// Add routing_mode and subdomain fields to projects collection
		projects.Fields.Add(&core.SelectField{
			Name:      "routing_mode",
			Required:  false,
			Values:    []string{"subdomain", "path"},
			MaxSelect: 1,
		})
		projects.Fields.Add(&core.TextField{
			Name:     "subdomain",
			Required: false,
			Min:      1,
			Max:      63,
			Pattern:  "^[a-z0-9]([a-z0-9-]*[a-z0-9])?$",
		})
		projects.Fields.Add(&core.RelationField{
			Name:          "owner_id",
			Required:      false,
			CollectionId:  "_pb_users_auth_",
			MaxSelect:     1,
			CascadeDelete: false,
		})

		// Add unique index for subdomain
		projects.AddIndex("idx_projects_subdomain", true, "subdomain", "subdomain IS NOT NULL")

		if err := app.Save(projects); err != nil {
			return err
		}

		// Create domains collection for custom domain mapping
		domains := core.NewBaseCollection("domains")
		domains.ListRule = ptrStr("@request.auth.id != ''")
		domains.ViewRule = ptrStr("@request.auth.id != ''")
		domains.CreateRule = ptrStr("@request.auth.id != ''")
		domains.UpdateRule = ptrStr("@request.auth.id != ''")
		domains.DeleteRule = ptrStr("@request.auth.id != ''")

		domains.Fields.Add(&core.TextField{
			Name:     "domain",
			Required: true,
			Min:      1,
			Max:      253,
		})
		domains.Fields.Add(&core.TextField{
			Name:     "slug",
			Required: true,
			Min:      1,
			Max:      255,
		})
		domains.Fields.Add(&core.RelationField{
			Name:          "project_id",
			Required:      true,
			CollectionId:  projects.Id,
			MaxSelect:     1,
			CascadeDelete: true,
		})
		domains.Fields.Add(&core.SelectField{
			Name:      "type",
			Required:  true,
			Values:    []string{"system", "custom"},
			MaxSelect: 1,
		})
		domains.Fields.Add(&core.SelectField{
			Name:      "status",
			Required:  true,
			Values:    []string{"pending", "verified", "active"},
			MaxSelect: 1,
		})
		domains.Fields.Add(&core.TextField{
			Name:     "verification_token",
			Required: false,
		})
		domains.Fields.Add(&core.BoolField{
			Name:     "is_primary",
			Required: false,
		})

		domains.AddIndex("idx_domains_lookup", true, "domain, slug", "")
		domains.AddIndex("idx_domains_project", false, "project_id", "")
		domains.AddIndex("idx_domains_status", false, "status", "")

		if err := app.Save(domains); err != nil {
			return err
		}

		// Add is_anonymous and expires_at fields to users collection
		users, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}

		users.Fields.Add(&core.BoolField{
			Name:     "is_anonymous",
			Required: false,
		})
		users.Fields.Add(&core.DateField{
			Name:     "anonymous_expires_at",
			Required: false,
		})

		if err := app.Save(users); err != nil {
			return err
		}

		return nil
	}, func(app core.App) error {
		// Remove domains collection
		domains, err := app.FindCollectionByNameOrId("domains")
		if err == nil {
			if err := app.Delete(domains); err != nil {
				return err
			}
		}

		// Remove added fields from projects
		projects, err := app.FindCollectionByNameOrId("projects")
		if err == nil {
			projects.Fields.RemoveByName("routing_mode")
			projects.Fields.RemoveByName("subdomain")
			projects.Fields.RemoveByName("owner_id")
			if err := app.Save(projects); err != nil {
				return err
			}
		}

		// Remove added fields from users
		users, err := app.FindCollectionByNameOrId("users")
		if err == nil {
			users.Fields.RemoveByName("is_anonymous")
			users.Fields.RemoveByName("anonymous_expires_at")
			if err := app.Save(users); err != nil {
				return err
			}
		}

		return nil
	})
}
