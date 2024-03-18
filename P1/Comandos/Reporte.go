package Comandos

import (
	"P1/Structs"
	"fmt"
	"strings"
)

func ValidarDatosREPORT(context []string) {
	name := "" // Inicializar variables fuera del bucle
	ruta := "" // Inicializar variables fuera del bucle
	path := ""
	id := ""
	for i := 0; i < len(context); i++ {
		token := context[i]
		tk := strings.Split(token, "=")
		if Comparar(tk[0], "id") {
			id = tk[1]
		}
		if Comparar(tk[0], "name") {
			if Comparar(tk[1], "mbr") || Comparar(tk[1], "disk") {
				name = tk[1]
			} else {
				Error("REP", "El comando name debe tener valores específicos")
				return
			}
		}
		if Comparar(tk[0], "path") {
			path = tk[1]
		}
		if Comparar(tk[0], "ruta") {
			ruta = tk[1]
		}
	}
	if id == "" {
		Error("REP", "El comando requiere el parámetro id obligatoriamente")
		return
	}
	if name == "" {
		Error("REP", "El comando requiere el parámetro name obligatoriamente")
	}
	if ruta == "" {
		Error("REP", "El comando requiere el parámetro ruta obligatoriamente")
	}
	if path == "" {
		Error("REP", "El comando requiere el parámetro path obligatoriamente")
	}
	generarReporteMBR(id)
	// Ahora name y ruta están siendo utilizados en el alcance de la función
}

func generarReporteMBR(id string) {
	// Arreglo con todas las letras del abecedario en mayúscula

	// Recorrer cada letra del abecedario
	for _, letra := range alfabeto {

		// Construir el nombre del disco usando la letra actual
		path := string(letra) + ".dsk"

		// Obtener la partición correspondiente al ID en el disco actual
		particion := GetMount("REP", id, &path)

		// Verificar si se encontró la partición en el disco actual
		if particion != (Structs.Particion{}) {
			// Mostrar información de la partición encontrada
			fmt.Println("------- Reporte de Partición -------")
			mostrarInfoParticion(particion)
			return
		}
	}

	// Si no se encontró la partición en ningún disco, imprimir un mensaje de error
	fmt.Println("No se encontró una partición con el ID especificado en ningún disco.")
}

// Función mostrarInfoParticion y estructuras omitidas por brevedad

func mostrarInfoParticion(particion Structs.Particion) {
	fmt.Printf("Estado: %c\n", particion.Part_status)
	fmt.Printf("Tipo: %c\n", particion.Part_type)
	fmt.Printf("Ajuste: %c\n", particion.Part_fit)
	fmt.Printf("Inicio: %d\n", particion.Part_start)
	fmt.Printf("Tamaño: %d\n", particion.Part_size)
	fmt.Printf("Nombre: %s\n", string(particion.Part_name[:]))
}

func mostrarInfoEBR(ebr Structs.EBR) {
	fmt.Printf("Estado: %c\n", ebr.Part_status)
	fmt.Printf("Ajuste: %c\n", ebr.Part_fit)
	fmt.Printf("Inicio: %d\n", ebr.Part_start)
	fmt.Printf("Tamaño: %d\n", ebr.Part_size)
	fmt.Printf("Siguiente: %d\n", ebr.Part_next)
	fmt.Printf("Nombre: %s\n", string(ebr.Part_name[:]))
}
