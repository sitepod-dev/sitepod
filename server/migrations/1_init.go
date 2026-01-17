package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		// Create projects collection
		projects := core.NewBaseCollection("projects")
		projects.ListRule = ptrStr("@request.auth.id != ''")
		projects.ViewRule = ptrStr("@request.auth.id != ''")
		projects.CreateRule = ptrStr("@request.auth.id != ''")
		projects.UpdateRule = ptrStr("@request.auth.id != ''")
		projects.DeleteRule = ptrStr("@request.auth.id != ''")

		projects.Fields.Add(&core.TextField{
			Name:     "name",
			Required: true,
			Min:      1,
			Max:      100,
		})

		projects.AddIndex("idx_projects_name", true, "name", "")

		if err := app.Save(projects); err != nil {
			return err
		}

		// Create images collection
		images := core.NewBaseCollection("images")
		images.ListRule = ptrStr("@request.auth.id != ''")
		images.ViewRule = ptrStr("@request.auth.id != ''")
		images.CreateRule = ptrStr("@request.auth.id != ''")
		images.UpdateRule = nil
		images.DeleteRule = ptrStr("@request.auth.id != ''")

		images.Fields.Add(&core.TextField{
			Name:     "image_id",
			Required: true,
		})
		images.Fields.Add(&core.RelationField{
			Name:          "project_id",
			Required:      true,
			CollectionId:  projects.Id,
			MaxSelect:     1,
			CascadeDelete: false,
		})
		images.Fields.Add(&core.TextField{
			Name:     "content_hash",
			Required: true,
		})
		images.Fields.Add(&core.JSONField{
			Name:     "manifest",
			Required: true,
		})
		images.Fields.Add(&core.NumberField{
			Name:     "file_count",
			Required: false,
		})
		images.Fields.Add(&core.NumberField{
			Name:     "total_size",
			Required: false,
		})
		images.Fields.Add(&core.TextField{
			Name:     "git_commit",
			Required: false,
		})
		images.Fields.Add(&core.TextField{
			Name:     "git_branch",
			Required: false,
		})
		images.Fields.Add(&core.TextField{
			Name:     "git_message",
			Required: false,
		})

		images.AddIndex("idx_images_project", false, "project_id", "")
		images.AddIndex("idx_images_image_id", false, "image_id", "")
		images.AddIndex("idx_images_hash", false, "project_id, content_hash", "")

		if err := app.Save(images); err != nil {
			return err
		}

		// Create plans collection
		plans := core.NewBaseCollection("plans")
		plans.ListRule = ptrStr("@request.auth.id != ''")
		plans.ViewRule = ptrStr("@request.auth.id != ''")
		plans.CreateRule = ptrStr("@request.auth.id != ''")
		plans.UpdateRule = ptrStr("@request.auth.id != ''")
		plans.DeleteRule = ptrStr("@request.auth.id != ''")

		plans.Fields.Add(&core.TextField{
			Name:     "plan_id",
			Required: true,
		})
		plans.Fields.Add(&core.RelationField{
			Name:          "project_id",
			Required:      true,
			CollectionId:  projects.Id,
			MaxSelect:     1,
			CascadeDelete: false,
		})
		plans.Fields.Add(&core.TextField{
			Name:     "content_hash",
			Required: true,
		})
		plans.Fields.Add(&core.JSONField{
			Name:     "manifest",
			Required: true,
		})
		plans.Fields.Add(&core.JSONField{
			Name:     "missing_blobs",
			Required: false,
		})
		plans.Fields.Add(&core.SelectField{
			Name:      "upload_mode",
			Required:  true,
			Values:    []string{"presigned", "direct"},
			MaxSelect: 1,
		})
		plans.Fields.Add(&core.SelectField{
			Name:      "status",
			Required:  true,
			Values:    []string{"pending", "committed", "expired"},
			MaxSelect: 1,
		})
		plans.Fields.Add(&core.DateField{
			Name:     "expires_at",
			Required: true,
		})
		plans.Fields.Add(&core.TextField{
			Name:     "git_commit",
			Required: false,
		})
		plans.Fields.Add(&core.TextField{
			Name:     "git_branch",
			Required: false,
		})
		plans.Fields.Add(&core.TextField{
			Name:     "git_message",
			Required: false,
		})

		plans.AddIndex("idx_plans_plan_id", false, "plan_id", "")
		plans.AddIndex("idx_plans_status", false, "status, expires_at", "")

		if err := app.Save(plans); err != nil {
			return err
		}

		// Create deploy_events collection
		deployEvents := core.NewBaseCollection("deploy_events")
		deployEvents.ListRule = ptrStr("@request.auth.id != ''")
		deployEvents.ViewRule = ptrStr("@request.auth.id != ''")
		deployEvents.CreateRule = nil
		deployEvents.UpdateRule = nil
		deployEvents.DeleteRule = nil

		deployEvents.Fields.Add(&core.RelationField{
			Name:          "project_id",
			Required:      true,
			CollectionId:  projects.Id,
			MaxSelect:     1,
			CascadeDelete: false,
		})
		deployEvents.Fields.Add(&core.RelationField{
			Name:          "image_id",
			Required:      true,
			CollectionId:  images.Id,
			MaxSelect:     1,
			CascadeDelete: false,
		})
		deployEvents.Fields.Add(&core.SelectField{
			Name:      "environment",
			Required:  true,
			Values:    []string{"prod", "beta"},
			MaxSelect: 1,
		})
		deployEvents.Fields.Add(&core.SelectField{
			Name:      "action",
			Required:  true,
			Values:    []string{"deploy", "rollback"},
			MaxSelect: 1,
		})
		deployEvents.Fields.Add(&core.TextField{
			Name:     "previous_image_id",
			Required: false,
		})

		deployEvents.AddIndex("idx_deploy_events_lookup", false, "project_id, environment", "")

		if err := app.Save(deployEvents); err != nil {
			return err
		}

		// Create previews collection
		previews := core.NewBaseCollection("previews")
		previews.ListRule = ptrStr("@request.auth.id != ''")
		previews.ViewRule = ptrStr("@request.auth.id != ''")
		previews.CreateRule = ptrStr("@request.auth.id != ''")
		previews.UpdateRule = nil
		previews.DeleteRule = ptrStr("@request.auth.id != ''")

		previews.Fields.Add(&core.TextField{
			Name:     "project",
			Required: true,
		})
		previews.Fields.Add(&core.RelationField{
			Name:          "image_id",
			Required:      true,
			CollectionId:  images.Id,
			MaxSelect:     1,
			CascadeDelete: false,
		})
		previews.Fields.Add(&core.TextField{
			Name:     "slug",
			Required: true,
		})
		previews.Fields.Add(&core.DateField{
			Name:     "expires_at",
			Required: true,
		})

		previews.AddIndex("idx_previews_slug", true, "project, slug", "")
		previews.AddIndex("idx_previews_expires", false, "expires_at", "")

		if err := app.Save(previews); err != nil {
			return err
		}

		return nil
	}, func(app core.App) error {
		// Rollback: drop all collections
		collections := []string{"previews", "deploy_events", "plans", "images", "projects"}
		for _, name := range collections {
			collection, err := app.FindCollectionByNameOrId(name)
			if err == nil {
				if err := app.Delete(collection); err != nil {
					return err
				}
			}
		}

		return nil
	})
}

func ptrStr(s string) *string {
	return &s
}
