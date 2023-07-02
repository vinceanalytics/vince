package models

import (
	"context"
	"net/url"

	"github.com/google/uuid"
	"github.com/vinceanalytics/vince/pkg/schema"
	"gorm.io/gorm"
)

type Invitation = schema.Invitation

func NewInvite(ctx context.Context, i *Invitation) (o url.Values) {
	o = make(url.Values)
	i.UUID = uuid.New()
	if i.Email == "" {
		o.Add("email", "email is required")
	}
	if i.Role == "" {
		o.Add("role", "role is required")
	}
	if len(o) > 0 {
		return
	}
	exists := Exists(ctx, func(db *gorm.DB) *gorm.DB {
		return db.Model(&Invitation{}).Where("email = ?", i.Email).Where("site_id = ?", i.SiteID)
	})
	if exists {
		o.Set("invite", "invitation has already been sent")
		return
	}
	return
}
