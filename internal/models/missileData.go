package models

type WeaponData struct {
	WeaponFileName   string
	WeaponName       string
	WeaponType       string
	TargetType       string
	WeaponProperties WeaponProperties
}
type WeaponProperties struct {
	MaxVelocity    string
	MinLaunchRange string
	MaxLaunchRange string
}
