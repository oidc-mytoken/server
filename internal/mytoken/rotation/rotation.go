package rotation

import (
	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/api/v0"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/encryptionkeyrepo"
	helper "github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo/mytokenrepohelper"
	"github.com/oidc-mytoken/server/internal/endpoints/token/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/model"
	eventService "github.com/oidc-mytoken/server/internal/mytoken/event"
	pkg2 "github.com/oidc-mytoken/server/internal/mytoken/event/pkg"
	mytoken "github.com/oidc-mytoken/server/internal/mytoken/pkg"
)

func rotateMytoken(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, oldJWT string, old *mytoken.Mytoken, clientMetaData api.ClientMetaData,
) (*mytoken.Mytoken, bool, error) {
	rotated := old.Rotate()
	jwt, err := rotated.ToJWT()
	if err != nil {
		return old, false, err
	}
	if err = db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			if err = helper.UpdateSeqNo(rlog, tx, rotated.ID, rotated.SeqNo); err != nil {
				return err
			}
			if err = encryptionkeyrepo.ReencryptEncryptionKey(rlog, tx, rotated.ID, oldJWT, jwt); err != nil {
				return err
			}
			return eventService.LogEvent(
				rlog, tx, pkg2.MTEvent{
					Event:          api.EventMTRotated,
					MTID:           rotated.ID,
					ClientMetaData: clientMetaData,
				},
			)
		},
	); err != nil {
		return old, false, err
	}
	return rotated, true, nil
}

// rotateMytokenAfterAT rotates a mytoken after it was used to obtain an AT if rotation is enabled for that case
func rotateMytokenAfterAT(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, oldJWT string, old *mytoken.Mytoken, clientMetaData api.ClientMetaData,
) (*mytoken.Mytoken, bool, error) {
	if old.Rotation == nil {
		return old, false, nil
	}
	if !old.Rotation.OnAT {
		return old, false, nil
	}
	return rotateMytoken(rlog, tx, oldJWT, old, clientMetaData)
}

// rotateMytokenAfterOther rotates a mytoken after it was used for other usages than AT if rotation is enabled for that
// case
func rotateMytokenAfterOther(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, oldJWT string, old *mytoken.Mytoken, clientMetaData api.ClientMetaData,
) (*mytoken.Mytoken, bool, error) {
	if old.Rotation == nil {
		return old, false, nil
	}
	if !old.Rotation.OnOther {
		return old, false, nil
	}
	return rotateMytoken(rlog, tx, oldJWT, old, clientMetaData)
}

// RotateMytokenAfterOtherForResponse rotates a mytoken after it was used for other usages than AT if rotation is
// enabled for that case and returns a pkg.MytokenResponse with the updated infos
func RotateMytokenAfterOtherForResponse(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, oldJWT string, old *mytoken.Mytoken, clientMetaData api.ClientMetaData,
	responseType model.ResponseType,
) (*pkg.MytokenResponse, error) {
	my, rotated, err := rotateMytokenAfterOther(rlog, tx, oldJWT, old, clientMetaData)
	if err != nil {
		return nil, err
	}
	if !rotated {
		return nil, nil
	}
	resp, err := my.ToTokenResponse(rlog, responseType, 0, clientMetaData, "")
	return &resp, err
}

// RotateMytokenAfterATForResponse rotates a mytoken after it was used for obtaining an AT if rotation is enabled for
// that case and returns a pkg.MytokenResponse with the updated infos
func RotateMytokenAfterATForResponse(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, oldJWT string, old *mytoken.Mytoken, clientMetaData api.ClientMetaData,
	responseType model.ResponseType,
) (*pkg.MytokenResponse, error) {
	my, rotated, err := rotateMytokenAfterAT(rlog, tx, oldJWT, old, clientMetaData)
	if err != nil {
		return nil, err
	}
	if !rotated {
		return nil, nil
	}
	resp, err := my.ToTokenResponse(rlog, responseType, 0, clientMetaData, "")
	return &resp, err
}
