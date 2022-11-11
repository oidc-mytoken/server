package utils

import (
	"github.com/oidc-mytoken/server/internal/utils/hashutils"
	"github.com/oidc-mytoken/server/shared/utils/issuerutils"
)

// CreateMytokenSubject creates the subject of a Mytoken from the oidc subject and oidc issuer
func CreateMytokenSubject(oidcSub, oidcIss string) string {
	comb := issuerutils.CombineSubIss(oidcSub, oidcIss)
	return hashutils.SHA3_256Str([]byte(comb))
}
