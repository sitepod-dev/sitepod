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

		// Get projects collection for relation
		projects, err := dao.FindCollectionByNameOrId("projects")
		if err != nil {
			return err
		}

		// Add routing_mode and subdomain fields to projects collection
		projects.Schema.AddField(&schema.SchemaField{
			Name:     "routing_mode",
			Type:     schema.FieldTypeSelect,
			Required: false,
			Options: &schema.SelectOptions{
				Values:    []string{"subdomain", "path"},
				MaxSelect: 1,
			},
		})
		projects.Schema.AddField(&schema.SchemaField{
			Name:     "subdomain",
			Type:     schema.FieldTypeText,
			Required: false,
			Options: &schema.TextOptions{
				Min:     types.Pointer(1),
				Max:     types.Pointer(63),
				Pattern: "^[a-z0-9]([a-z0-9-]*[a-z0-9])?$",
			},
		})
		projects.Schema.AddField(&schema.SchemaField{
			Name:     "owner_id",
			Type:     schema.FieldTypeRelation,
			Required: false,
			Options: &schema.RelationOptions{
				CollectionId:  "_pb_users_auth_",
				MaxSelect:     types.Pointer(1),
				CascadeDelete: false,
			},
		})

		// Add unique index for subdomain
		projects.Indexes = append(projects.Indexes, "CREATE UNIQUE INDEX idx_projects_subdomain ON projects (subdomain) WHERE subdomain IS NOT NULL")

		if err := dao.SaveCollection(projects); err != nil {
			return err
		}

		// Create domains collection for custom domain mapping
		domains := &models.Collection{
			Name:       "domains",
			Type:       models.CollectionTypeBase,
			ListRule:   types.Pointer("@request.auth.id != ''"),
			ViewRule:   types.Pointer("@request.auth.id != ''"),
			CreateRule: types.Pointer("@request.auth.id != ''"),
			UpdateRule: types.Pointer("@request.auth.id != ''"),
			DeleteRule: types.Pointer("@request.auth.id != ''"),
			Schema: schema.NewSchema(
				&schema.SchemaField{
					Name:     "domain",
					Type:     schema.FieldTypeText,
					Required: true,
					Options: &schema.TextOptions{
						Min: types.Pointer(1),
						Max: types.Pointer(253),
					},
				},
				&schema.SchemaField{
					Name:     "slug",
					Type:     schema.FieldTypeText,
					Required: true,
					Options: &schema.TextOptions{
						Min: types.Pointer(1),
						Max: types.Pointer(255),
					},
				},
				&schema.SchemaField{
					Name:     "project_id",
					Type:     schema.FieldTypeRelation,
					Required: true,
					Options: &schema.RelationOptions{
						CollectionId:  projects.Id,
						MaxSelect:     types.Pointer(1),
						CascadeDelete: true,
					},
				},
				&schema.SchemaField{
					Name:     "type",
					Type:     schema.FieldTypeSelect,
					Required: true,
					Options: &schema.SelectOptions{
						Values:    []string{"system", "custom"},
						MaxSelect: 1,
					},
				},
				&schema.SchemaField{
					Name:     "status",
					Type:     schema.FieldTypeSelect,
					Required: true,
					Options: &schema.SelectOptions{
						Values:    []string{"pending", "verified", "active"},
						MaxSelect: 1,
					},
				},
				&schema.SchemaField{
					Name:     "verification_token",
					Type:     schema.FieldTypeText,
					Required: false,
				},
				&schema.SchemaField{
					Name:     "is_primary",
					Type:     schema.FieldTypeBool,
					Required: false,
				},
			),
			Indexes: types.JsonArray[string]{
				"CREATE UNIQUE INDEX idx_domains_lookup ON domains (domain, slug)",
				"CREATE INDEX idx_domains_project ON domains (project_id)",
				"CREATE INDEX idx_domains_status ON domains (status)",
			},
		}
		if err := dao.SaveCollection(domains); err != nil {
			return err
		}

		// Add is_anonymous and expires_at fields to users collection
		users, err := dao.FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}

		users.Schema.AddField(&schema.SchemaField{
			Name:     "is_anonymous",
			Type:     schema.FieldTypeBool,
			Required: false,
		})
		users.Schema.AddField(&schema.SchemaField{
			Name:     "anonymous_expires_at",
			Type:     schema.FieldTypeDate,
			Required: false,
		})

		if err := dao.SaveCollection(users); err != nil {
			return err
		}

		return nil
	}, func(db dbx.Builder) error {
		dao := daos.New(db)

		// Remove domains collection
		domains, err := dao.FindCollectionByNameOrId("domains")
		if err == nil {
			if err := dao.DeleteCollection(domains); err != nil {
				return err
			}
		}

		// Remove added fields from projects
		projects, err := dao.FindCollectionByNameOrId("projects")
		if err == nil {
			projects.Schema.RemoveField("routing_mode")
			projects.Schema.RemoveField("subdomain")
			projects.Schema.RemoveField("owner_id")
			if err := dao.SaveCollection(projects); err != nil {
				return err
			}
		}

		// Remove added fields from users
		users, err := dao.FindCollectionByNameOrId("users")
		if err == nil {
			users.Schema.RemoveField("is_anonymous")
			users.Schema.RemoveField("anonymous_expires_at")
			if err := dao.SaveCollection(users); err != nil {
				return err
			}
		}

		return nil
	})
}
