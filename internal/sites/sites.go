package sites

import (
	"context"

	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/pkg/log"
)

func Role(ctx context.Context, userId uint64, siteId uint64) (role string) {
	err := models.Get(ctx).Model(&models.SiteMembership{}).
		Where("site_id=?", siteId).
		Where("user_id=?", userId).
		Select("role").Limit(1).Find(&role).Error
	if err != nil {
		log.Get().Err(err).Uint64("site_id", siteId).
			Uint64("user_id", userId).Msg("failed to get role for site membership")
	}
	return
}
