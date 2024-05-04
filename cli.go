package grua

import (
	"bytes"
	"os"
	"strings"
	"time"
	"unicode"
)

func NuevaMigracion(nombreMigracion string, rutaMigracion string) error {
	plantilla, err := os.ReadFile("plantillaMigracion.txt")
	if err != nil {
		return err
	}

	nombreConMarcaDeTiempo := obtenerMarcaDeTiempo() + camelCaseASnakeCase(nombreMigracion)
	rutaMigracionDirectorios := strings.Split(rutaMigracion, "/")
	nombrePaqueteMigracion := rutaMigracionDirectorios[len(rutaMigracionDirectorios)-1]

	plantillaFormateada := strings.ReplaceAll(string(plantilla), "NombrePaquete", nombrePaqueteMigracion)
	plantillaFormateada = strings.ReplaceAll(plantillaFormateada, "NombreMigracion", nombreMigracion)
	plantillaFormateada = strings.ReplaceAll(plantillaFormateada, "NombreConMarcaDeTiempoMigracion", nombreConMarcaDeTiempo)

	rutaArchivoMigracion := rutaMigracion + string("/") + nombreConMarcaDeTiempo + string(".go")

	if _, err := os.Stat(rutaMigracion); err != nil {
		err := os.MkdirAll(rutaMigracion, 0755)
		if err != nil {
			return err
		}
	}

	err = os.WriteFile(rutaArchivoMigracion, []byte(plantillaFormateada), 0644)
	if err != nil {
		return err
	}

	return nil
}

func camelCaseASnakeCase(string string) string {
	var buffer bytes.Buffer
	var anterior rune

	for i, caracter := range string {
		if i > 0 && unicode.IsUpper(caracter) && !unicode.IsUpper(anterior) {
			buffer.WriteRune('_')
		}

		buffer.WriteRune(unicode.ToLower(caracter))
		anterior = caracter
	}

	return buffer.String()
}

func obtenerMarcaDeTiempo() string {
	fechaYHoraActual := time.Now()

	marcaDeTiempo := fechaYHoraActual.Format("2006_01_02_150405")

	return marcaDeTiempo + string('_')
}
