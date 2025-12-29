package data

import (
	"bufio"
	"io/fs"
	"log"
	"os"
	"strings"

	"github.com/stevenzieske/SeaPower-Calculator/configs"
	"github.com/stevenzieske/SeaPower-Calculator/internal/models"
)

var cfg = configs.Load()

func LoadMissileData() []models.WeaponData {
	originalDirectory := cfg.GAME_DIR + "/Sea Power/Sea Power_Data/StreamingAssets/original/"

	ammunitionFS := os.DirFS(originalDirectory + "ammunition/")

	ammunitionFiles, err := fs.Glob(ammunitionFS, "*.ini")

	if err != nil {
		log.Fatal(err)
	}

	var weapons []models.WeaponData
	for _, fileName := range ammunitionFiles {
		weapons = append(weapons, ExtractWeaponDataFromFile(fileName, originalDirectory))
	}

	return weapons
}

func ExtractWeaponDataFromFile(fileName string, originalDirectory string) models.WeaponData {
	weapon := models.WeaponData{
		WeaponFileName: fileName,
	}
	file, err := os.Open(originalDirectory + "ammunition/" + weapon.WeaponFileName)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	r := bufio.NewReader(file)

	for {
		line, _, err := r.ReadLine()
		if len(line) > 0 {
			cleanLine := strings.TrimSpace(strings.Split(string(line), "//")[0])

			if strings.Contains(cleanLine, "=") {
				parts := strings.SplitN(cleanLine, "=", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					value := strings.TrimSpace(parts[1])

					switch strings.ToLower(key) {
					case "type":
						weapon.WeaponType = value
					case "targettype":
						weapon.TargetType = value
					case "maxvelocity":
						weapon.WeaponProperties.MaxVelocity = value
					case "minlaunchrange":
						weapon.WeaponProperties.MinLaunchRange = value
					case "maxlaunchrange":
						weapon.WeaponProperties.MaxLaunchRange = value
					}
				}
			}
		}
		if err != nil {
			break
		}
	}
	file.Close()

	file, err = os.Open(originalDirectory + "language_en/ammunition_names.ini")
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}
	defer file.Close()
	r = bufio.NewReader(file)

	baseFileName := strings.TrimSuffix(weapon.WeaponFileName, ".ini")

	for {
		line, _, err := r.ReadLine()
		if len(line) > 0 && strings.Contains(string(line), baseFileName) {
			cleanLine := strings.TrimSpace(string(line))

			if strings.Contains(cleanLine, "=") {
				parts := strings.SplitN(cleanLine, "=", 2)
				if len(parts) == 2 {
					value := strings.TrimSpace(parts[1])

					nameParts := strings.Split(value, ",")
					if len(nameParts) > 0 {
						weapon.WeaponName = strings.TrimSpace(nameParts[0])
						if len(nameParts) > 1 && strings.TrimSpace(nameParts[1]) != "" {
							weapon.WeaponName += " - " + strings.TrimSpace(nameParts[1])
						}
						if len(nameParts) > 2 && strings.TrimSpace(nameParts[2]) != "" {
							weapon.WeaponName += " (" + strings.TrimSpace(nameParts[2]) + ")"
						}
					}
				}
			}
		}
		if err != nil {
			break
		}
	}

	return weapon
}
