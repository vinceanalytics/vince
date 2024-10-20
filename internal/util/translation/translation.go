package translation

import "github.com/vinceanalytics/vince/internal/models"

type Translator interface {
	Translate(field models.Field, value []byte) uint64
}
