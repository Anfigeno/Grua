package grua

import (
	"fmt"

	"gorm.io/gorm"
)

type ManejadorDeMigraciones struct {
	bd                   *gorm.DB
	migracionesAplicadas map[string]migracionCompletada
	migraciones          []Migracion
}

type Migracion struct {
	ID          string
	Descripcion string
	Aplicar     func(bd *gorm.DB) error
	Revertir    func(bd *gorm.DB) error
}

type migracionCompletada struct {
	gorm.Model
	IdMigracion string
	Lote        int
}

func NuevoManejadorDeMigraciones(bd *gorm.DB) *ManejadorDeMigraciones {
	if !bd.Migrator().HasTable(&migracionCompletada{}) {
		bd.AutoMigrate(&migracionCompletada{})
	}

	var migracionesAplicadas []migracionCompletada
	bd.Find(&migracionesAplicadas)

	migracionesAplicadasMap := make(map[string]migracionCompletada)
	for _, migracionAplicada := range migracionesAplicadas {
		migracionesAplicadasMap[migracionAplicada.IdMigracion] = migracionAplicada
	}

	return &ManejadorDeMigraciones{
		bd:                   bd,
		migracionesAplicadas: migracionesAplicadasMap,
		migraciones:          []Migracion{},
	}
}

func (m *ManejadorDeMigraciones) AñadirMigraciones(migraciones ...Migracion) {
	for _, migracion := range migraciones {
		m.migraciones = append(m.migraciones, migracion)
	}
}

func obtenerLoteActual(migracionesAplicadas map[string]migracionCompletada) int {
	var loteActual int
	for _, migracionAplicada := range migracionesAplicadas {
		if migracionAplicada.Lote > loteActual {
			loteActual = migracionAplicada.Lote
		}
	}
	return loteActual
}

func (m *ManejadorDeMigraciones) AplicarMigraciones(registrador func(registro string)) error {
	var nuevasMigracionesAplicadas []migracionCompletada
	loteActual := obtenerLoteActual(m.migracionesAplicadas)

	registro := fmt.Sprintf("Aplicando el nuevo lote de migraciones %d...", loteActual+1)
	registrador(registro)

	for _, migracion := range m.migraciones {
		if _, ok := m.migracionesAplicadas[migracion.ID]; ok {
			continue
		}

		registro = fmt.Sprintf("Aplicando la migracion %s: %s", migracion.ID, migracion.Descripcion)
		registrador(registro)
		if err := migracion.Aplicar(m.bd); err != nil {
			registro = fmt.Sprintf("Error al aplicar la migración %s", migracion.ID)
			registrador(registro)

			return err
		}

		nuevaMigracionAplicada := migracionCompletada{
			IdMigracion: migracion.ID,
			Lote:        loteActual + 1,
		}

		nuevasMigracionesAplicadas = append(nuevasMigracionesAplicadas, nuevaMigracionAplicada)
		m.migracionesAplicadas[migracion.ID] = nuevaMigracionAplicada
	}

	if len(nuevasMigracionesAplicadas) == 0 {
		registro = "No hay nada que hacer"
		registrador(registro)

		return nil
	}

	registro = "Actualizando la tabla de migraciones..."
	registrador(registro)

	resultado := m.bd.Create(nuevasMigracionesAplicadas)
	if resultado.Error != nil {
		registro := "Error al actualizar la tabla de migraciones"
		registrador(registro)

		return resultado.Error
	}

	return nil
}

func (m *ManejadorDeMigraciones) RevertirMigraciones(registrador func(registro string)) error {
	if len(m.migracionesAplicadas) == 0 {
		registro := "No hay nada que hacer"
		registrador(registro)

		return nil
	}

	ultimoLote := obtenerLoteActual(m.migracionesAplicadas)

	registro := fmt.Sprintf("Revirtiendo el ultimo lote de migraciones %d...", ultimoLote)
	registrador(registro)

	var migracionesRevertidas []migracionCompletada

	for _, migracion := range m.migraciones {
		if m.migracionesAplicadas[migracion.ID].Lote != ultimoLote {
			continue
		}

		registro = fmt.Sprintf("Revirtiendo la migración %s: %s", migracion.ID, migracion.Descripcion)
		registrador(registro)

		if err := migracion.Revertir(m.bd); err != nil {
			registro = fmt.Sprintf("Error al revertir la migración %s", migracion.ID)
			return err
		}

		migracionesRevertidas = append(migracionesRevertidas, m.migracionesAplicadas[migracion.ID])
		delete(m.migracionesAplicadas, migracion.ID)
	}

	if len(migracionesRevertidas) == 0 {
		registro = "No hay nada que hacer"
		registrador(registro)

		return nil
	}

	registro = "Actualizando la tabla de migraciones..."
	registrador(registro)

	for _, migracionRevertida := range migracionesRevertidas {
		resultado := m.bd.Where("id_migracion = ?", migracionRevertida.IdMigracion).Delete(&migracionRevertida)
		if resultado.Error != nil {
			registro = fmt.Sprint("Error al actualizar la tabla de migraciones")
			registrador(registro)

			return resultado.Error
		}
	}

	return nil
}
