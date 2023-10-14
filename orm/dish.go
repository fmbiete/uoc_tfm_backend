package orm

import (
	"errors"

	"gorm.io/gorm"
)

func (d *Database) DishCreate(dish Dish) (Dish, error) {
	err := d.db.Where("name = ?", dish.Name).First(&Dish{}).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err := d.db.Create(&dish).Error
		return dish, err
	}

	if err != nil {
		return dish, err
	}

	// No error, we have found a matching dish - return duplicated error
	return dish, gorm.ErrDuplicatedKey
}

func (d *Database) DishDelete(dishId uint64) error {
	return d.db.Delete(&Dish{}, dishId).Error
}

func (d *Database) DishDetails(dishId uint64) (Dish, error) {
	var dish Dish
	err := d.db.Preload("Allergens").Preload("Ingredients").First(&dish, dishId).Error
	return dish, err
}

func (d *Database) DishList(userId int64) ([]Dish, error) {
	var dishes []Dish
	// TODO: filter by userId if != -1, order by user sales
	// TODO: order by global sales
	err := d.db.Preload("Allergens").Find(&dishes).Error
	return dishes, err
}

func (d *Database) DishModify(dish Dish) (Dish, error) {
	// replace alergenos and ingredientes - Update adds new records, but doesn't delete old ones
	alergenos := dish.Allergens
	d.db.Unscoped().Model(&dish).Association("Allergens").Unscoped().Clear()
	dish.Allergens = alergenos

	ingredientes := dish.Ingredients
	d.db.Unscoped().Model(&dish).Association("Ingredients").Unscoped().Clear()
	dish.Ingredients = ingredientes

	err := d.db.Updates(&dish).Error
	if err != nil {
		return dish, err
	}

	return d.DishDetails(uint64(dish.ID))
}

func (d *Database) dishCurrentCost(dishId uint64) (float64, error) {
	var err error
	var dish Dish

	err = d.db.Select("cost").First(&dish, dishId).Error
	return dish.Cost, err

	// TODO: promociones
}
