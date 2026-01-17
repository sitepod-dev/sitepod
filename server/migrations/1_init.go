package migrations

import (
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/models/schema"
	"github.com/pocketbase/pocketbase/tools/types"
)

func init() {
	m.Register(func(db dbx.Builder) error {
		dao := daos.New(db)

		// Create projects collection
		projects := &models.Collection{
			Name:       "projects",
			Type:       models.CollectionTypeBase,
			ListRule:   types.Pointer("@request.auth.id != ''"),
			ViewRule:   types.Pointer("@request.auth.id != ''"),
			CreateRule: types.Pointer("@request.auth.id != ''"),
			UpdateRule: types.Pointer("@request.auth.id != ''"),
			DeleteRule: types.Pointer("@request.auth.id != ''"),
			Schema: schema.NewSchema(
				&schema.SchemaField{
					Name:     "name",
					Type:     schema.FieldTypeText,
					Required: true,
					Options: &schema.TextOptions{
						Min: types.Pointer(1),
						Max: types.Pointer(100),
					},
				},
			),
			Indexes: types.JsonArray[string]{
				"CREATE UNIQUE INDEX idx_projects_name ON projects (name)",
			},
		}
		if err := dao.SaveCollection(projects); err != nil {
			return err
		}

		// Create images collection
		images := &models.Collection{
			Name:       "images",
			Type:       models.CollectionTypeBase,
			ListRule:   types.Pointer("@request.auth.id != ''"),
			ViewRule:   types.Pointer("@request.auth.id != ''"),
			CreateRule: types.Pointer("@request.auth.id != ''"),
			UpdateRule: nil,
			DeleteRule: types.Pointer("@request.auth.id != ''"),
			Schema: schema.NewSchema(
				&schema.SchemaField{
					Name:     "image_id",
					Type:     schema.FieldTypeText,
					Required: true,
				},
				&schema.SchemaField{
					Name:     "project_id",
					Type:     schema.FieldTypeRelation,
					Required: true,
					Options: &schema.RelationOptions{
						CollectionId:  projects.Id,
						MaxSelect:     types.Pointer(1),
						CascadeDelete: false,
					},
				},
				&schema.SchemaField{
					Name:     "content_hash",
					Type:     schema.FieldTypeText,
					Required: true,
				},
				&schema.SchemaField{
					Name:     "manifest",
					Type:     schema.FieldTypeJson,
					Required: true,
				},
				&schema.SchemaField{
					Name:     "file_count",
					Type:     schema.FieldTypeNumber,
					Required: false,
				},
				&schema.SchemaField{
					Name:     "total_size",
					Type:     schema.FieldTypeNumber,
					Required: false,
				},
				&schema.SchemaField{
					Name:     "git_commit",
					Type:     schema.FieldTypeText,
					Required: false,
				},
				&schema.SchemaField{
					Name:     "git_branch",
					Type:     schema.FieldTypeText,
					Required: false,
				},
				&schema.SchemaField{
					Name:     "git_message",
					Type:     schema.FieldTypeText,
					Required: false,
				},
			),
			Indexes: types.JsonArray[string]{
				"CREATE INDEX idx_images_project ON images (project_id)",
				"CREATE INDEX idx_images_image_id ON images (image_id)",
				"CREATE INDEX idx_images_hash ON images (project_id, content_hash)",
			},
		}
		if err := dao.SaveCollection(images); err != nil {
			return err
		}

		// Create plans collection
		plans := &models.Collection{
			Name:       "plans",
			Type:       models.CollectionTypeBase,
			ListRule:   types.Pointer("@request.auth.id != ''"),
			ViewRule:   types.Pointer("@request.auth.id != ''"),
			CreateRule: types.Pointer("@request.auth.id != ''"),
			UpdateRule: types.Pointer("@request.auth.id != ''"),
			DeleteRule: types.Pointer("@request.auth.id != ''"),
			Schema: schema.NewSchema(
				&schema.SchemaField{
					Name:     "plan_id",
					Type:     schema.FieldTypeText,
					Required: true,
				},
				&schema.SchemaField{
					Name:     "project_id",
					Type:     schema.FieldTypeRelation,
					Required: true,
					Options: &schema.RelationOptions{
						CollectionId:  projects.Id,
						MaxSelect:     types.Pointer(1),
						CascadeDelete: false,
					},
				},
				&schema.SchemaField{
					Name:     "content_hash",
					Type:     schema.FieldTypeText,
					Required: true,
				},
				&schema.SchemaField{
					Name:     "manifest",
					Type:     schema.FieldTypeJson,
					Required: true,
				},
				&schema.SchemaField{
					Name:     "missing_blobs",
					Type:     schema.FieldTypeJson,
					Required: false,
				},
				&schema.SchemaField{
					Name:     "upload_mode",
					Type:     schema.FieldTypeSelect,
					Required: true,
					Options: &schema.SelectOptions{
						Values:    []string{"presigned", "direct"},
						MaxSelect: 1,
					},
				},
				&schema.SchemaField{
					Name:     "status",
					Type:     schema.FieldTypeSelect,
					Required: true,
					Options: &schema.SelectOptions{
						Values:    []string{"pending", "committed", "expired"},
						MaxSelect: 1,
					},
				},
				&schema.SchemaField{
					Name:     "expires_at",
					Type:     schema.FieldTypeDate,
					Required: true,
				},
				&schema.SchemaField{
					Name:     "git_commit",
					Type:     schema.FieldTypeText,
					Required: false,
				},
				&schema.SchemaField{
					Name:     "git_branch",
					Type:     schema.FieldTypeText,
					Required: false,
				},
				&schema.SchemaField{
					Name:     "git_message",
					Type:     schema.FieldTypeText,
					Required: false,
				},
			),
			Indexes: types.JsonArray[string]{
				"CREATE INDEX idx_plans_plan_id ON plans (plan_id)",
				"CREATE INDEX idx_plans_status ON plans (status, expires_at)",
			},
		}
		if err := dao.SaveCollection(plans); err != nil {
			return err
		}

		// Create deploy_events collection
		deployEvents := &models.Collection{
			Name:       "deploy_events",
			Type:       models.CollectionTypeBase,
			ListRule:   types.Pointer("@request.auth.id != ''"),
			ViewRule:   types.Pointer("@request.auth.id != ''"),
			CreateRule: nil,
			UpdateRule: nil,
			DeleteRule: nil,
			Schema: schema.NewSchema(
				&schema.SchemaField{
					Name:     "project_id",
					Type:     schema.FieldTypeRelation,
					Required: true,
					Options: &schema.RelationOptions{
						CollectionId:  projects.Id,
						MaxSelect:     types.Pointer(1),
						CascadeDelete: false,
					},
				},
				&schema.SchemaField{
					Name:     "image_id",
					Type:     schema.FieldTypeRelation,
					Required: true,
					Options: &schema.RelationOptions{
						CollectionId:  images.Id,
						MaxSelect:     types.Pointer(1),
						CascadeDelete: false,
					},
				},
				&schema.SchemaField{
					Name:     "environment",
					Type:     schema.FieldTypeSelect,
					Required: true,
					Options: &schema.SelectOptions{
						Values:    []string{"prod", "beta"},
						MaxSelect: 1,
					},
				},
				&schema.SchemaField{
					Name:     "action",
					Type:     schema.FieldTypeSelect,
					Required: true,
					Options: &schema.SelectOptions{
						Values:    []string{"deploy", "rollback"},
						MaxSelect: 1,
					},
				},
				&schema.SchemaField{
					Name:     "previous_image_id",
					Type:     schema.FieldTypeText,
					Required: false,
				},
			),
			Indexes: types.JsonArray[string]{
				"CREATE INDEX idx_deploy_events_lookup ON deploy_events (project_id, environment)",
			},
		}
		if err := dao.SaveCollection(deployEvents); err != nil {
			return err
		}

		// Create previews collection
		previews := &models.Collection{
			Name:       "previews",
			Type:       models.CollectionTypeBase,
			ListRule:   types.Pointer("@request.auth.id != ''"),
			ViewRule:   types.Pointer("@request.auth.id != ''"),
			CreateRule: types.Pointer("@request.auth.id != ''"),
			UpdateRule: nil,
			DeleteRule: types.Pointer("@request.auth.id != ''"),
			Schema: schema.NewSchema(
				&schema.SchemaField{
					Name:     "project",
					Type:     schema.FieldTypeText,
					Required: true,
				},
				&schema.SchemaField{
					Name:     "image_id",
					Type:     schema.FieldTypeRelation,
					Required: true,
					Options: &schema.RelationOptions{
						CollectionId:  images.Id,
						MaxSelect:     types.Pointer(1),
						CascadeDelete: false,
					},
				},
				&schema.SchemaField{
					Name:     "slug",
					Type:     schema.FieldTypeText,
					Required: true,
				},
				&schema.SchemaField{
					Name:     "expires_at",
					Type:     schema.FieldTypeDate,
					Required: true,
				},
			),
			Indexes: types.JsonArray[string]{
				"CREATE UNIQUE INDEX idx_previews_slug ON previews (project, slug)",
				"CREATE INDEX idx_previews_expires ON previews (expires_at)",
			},
		}
		if err := dao.SaveCollection(previews); err != nil {
			return err
		}

		return nil
	}, func(db dbx.Builder) error {
		// Rollback: drop all collections
		dao := daos.New(db)

		collections := []string{"previews", "deploy_events", "plans", "images", "projects"}
		for _, name := range collections {
			collection, err := dao.FindCollectionByNameOrId(name)
			if err == nil {
				if err := dao.DeleteCollection(collection); err != nil {
					return err
				}
			}
		}

		return nil
	})
}
