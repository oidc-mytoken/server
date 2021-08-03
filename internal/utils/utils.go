package utils

import (
	"github.com/oidc-mytoken/server/internal/utils/hashUtils"
	"github.com/oidc-mytoken/server/shared/utils/issuerUtils"
)

// CreateMytokenSubject creates the subject of a Mytoken from the oidc subject and oidc issuer
func CreateMytokenSubject(oidcSub, oidcIss string) string {
	comb := issuerUtils.CombineSubIss(oidcSub, oidcIss)
	return hashUtils.SHA3_256Str([]byte(comb))
}
