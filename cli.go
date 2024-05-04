package grua

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
	"unicode"
)

func (m *ManejadorDeMigraciones) NuevaMigracion(nombreMigracion string) error {
	plantilla, err := os.ReadFile("plantillaMigracion.txt")
	if err != nil {
		return err
	}

	nombreConMarcaDeTiempo := obtenerMarcaDeTiempo() + camelCaseASnakeCase(nombreMigracion)
	rutaMigracionDirectorios := strings.Split(m.RutaMigraciones, "/")
	nombrePaqueteMigracion := rutaMigracionDirectorios[len(rutaMigracionDirectorios)-1]

	plantillaFormateada := strings.ReplaceAll(string(plantilla), "NombrePaquete", nombrePaqueteMigracion)
	plantillaFormateada = strings.ReplaceAll(plantillaFormateada, "NombreMigracion", nombreMigracion)
	plantillaFormateada = strings.ReplaceAll(plantillaFormateada, "NombreConMarcaDeTiempoMigracion", nombreConMarcaDeTiempo)

	rutaArchivoMigracion := m.RutaMigraciones + string("/") + nombreConMarcaDeTiempo + string(".go")

	registro := fmt.Sprint("Creando migracion ", nombreConMarcaDeTiempo, "...")
	m.Registrador(registro)

	if _, err := os.Stat(m.RutaMigraciones); err != nil {
		err := os.MkdirAll(m.RutaMigraciones, 0755)
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

func Registrador(registro string) {
	log.Print(registro)
}

func (m *ManejadorDeMigraciones) ManejarCli(argumentos []string) error {
	accion := argumentos[1]

	switch accion {
	case "nueva":
		if len(argumentos) < 3 {
			return errors.New("Se necesita un nombre para la migración")
		}

		nombreMigracion := argumentos[2]

		err := m.NuevaMigracion(nombreMigracion)
		if err != nil {
			return err
		}
		return nil

	case "migrar":
		err := m.AplicarMigraciones()
		if err != nil {
			return err
		}

	case "revertir":
		err := m.RevertirMigraciones()
		if err != nil {
			return err
		}
	}

	return nil
}
