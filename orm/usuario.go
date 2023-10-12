package orm

import (
	"errors"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

func (d *Database) UsuarioCrear(user Usuario) (Usuario, error) {

	err := d.db.Where("email = ?", user.Email).First(&Usuario{}).Error
	log.Debug().Err(err).Msg("Check create")
	if errors.Is(err, gorm.ErrRecordNotFound) {
		result := d.db.Create(&user)
		return user, result.Error
	}

	return user, gorm.ErrDuplicatedKey
}

func (d *Database) UsuarioEliminar(userId int64) error {
	result := d.db.Delete(&Usuario{}, userId)
	return result.Error
}

func (d *Database) UsuarioGet(userId int64) (Usuario, error) {
	var user Usuario
	result := d.db.First(&user, userId)
	// Don't return the password hash
	user.Password = ""
	return user, result.Error
}

func (d *Database) UsuarioModificar(user Usuario) (Usuario, error) {
	result := d.db.Updates(&user)
	// returns only modified fields
	if result.Error == nil {
		return d.UsuarioGet(int64(user.ID))
	}
	// Don't return the password hash
	user.Password = ""
	return user, result.Error
}
