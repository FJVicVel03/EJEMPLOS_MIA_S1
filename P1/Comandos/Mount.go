package Comandos

import (
	"P1/Structs"
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
	"unsafe"
)

var DiscMont [99]DiscoMontado

type DiscoMontado struct {
	Path        [150]byte
	Estado      byte
	Particiones [26]ParticionMontada
}

type ParticionMontada struct {
	Letra  byte
	Estado byte
	Nombre [20]byte
	ID     string // Agregar campo para el ID
}

var alfabeto = []byte{'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z'}
var lastLetterIndex int

func generarID() string {
	// Verificar si ya se han utilizado todas las letras del alfabeto
	if lastLetterIndex >= len(alfabeto) {
		return ""
	}

	// Generar el ID con la siguiente letra disponible y los últimos tres dígitos del carnet
	id := string(alfabeto[lastLetterIndex]) + "576"

	// Incrementar el índice de la última letra utilizada
	lastLetterIndex++

	return id
}
func ValidarDatosMOUNT(context []string) {
	name := ""
	driveLetter := ""
	for i := 0; i < len(context); i++ {
		current := context[i]
		comando := strings.Split(current, "=")
		if Comparar(comando[0], "name") {
			name = comando[1]
		} else if Comparar(comando[0], "driveletter") {
			driveLetter = comando[1]
		}
	}
	if driveLetter == "" || name == "" {
		Error("MOUNT", "El comando MOUNT requiere parámetros obligatorios")
		return
	}
	mount(driveLetter, name)
	listaMount()
}

func mount(driveLetter string, name string) {
	// Construir la ruta del disco
	path := driveLetter + ".dsk"

	// Abrir el archivo del disco
	file, err := os.Open(path)
	if err != nil {
		Error("MOUNT", "No se ha podido abrir el archivo del disco.")
		return
	}
	defer file.Close()

	// Leer el MBR del disco
	mbr := Structs.MBR{}
	err = binary.Read(file, binary.BigEndian, &mbr)
	if err != nil {
		Error("MOUNT", "Error al leer el MBR del disco.")
		return
	}

	// Buscar la partición por nombre
	particion := BuscarParticiones(mbr, name, driveLetter)
	if particion == nil {
		Error("MOUNT", "No se encontró la partición especificada en el disco.")
		return
	}

	// Verificar si la partición es extendida o lógica
	if particion.Part_type == 'E' || particion.Part_type == 'L' {
		Error("MOUNT", "No se puede montar una partición extendida o lógica.")
		return
	}

	// Generar un ID para la partición
	id := generarID()

	// Buscar una entrada disponible en la lista de montajes
	for i := 0; i < 99; i++ {
		if DiscMont[i].Estado == 0 {
			DiscMont[i].Estado = 1
			copy(DiscMont[i].Path[:], []byte(path))
			for j := 0; j < 26; j++ {
				if DiscMont[i].Particiones[j].Estado == 0 {
					DiscMont[i].Particiones[j].Estado = 1
					DiscMont[i].Particiones[j].Letra = alfabeto[j]
					copy(DiscMont[i].Particiones[j].Nombre[:], []byte(name))
					DiscMont[i].Particiones[j].ID = id // Guardar el ID
					Mensaje("MOUNT", "Se ha realizado correctamente el montaje de la partición. ID: "+id)
					return
				}
			}
		}
	}

	Error("MOUNT", "No hay espacio disponible para montar la partición.")
}

func ValidarDatosUNMOUNT(context []string) {
	id := ""
	for i := 0; i < len(context); i++ {
		current := context[i]
		comando := strings.Split(current, "=")
		if Comparar(comando[0], "id") {
			id = comando[1]
		}
	}
	if id == "" {
		Error("UNMOUNT", "El comando UNMOUNT requiere el parámetro obligatorio 'id'")
		return
	}
	unmount(id)
}

func unmount(id string) {
	// Buscar la partición correspondiente al ID en la lista de montajes
	for i := 0; i < 99; i++ {
		for j := 0; j < 26; j++ {
			if string(DiscMont[i].Particiones[j].Letra)+"576" == id {
				// Marcar la partición como desmontada
				DiscMont[i].Particiones[j].Estado = 0
				Mensaje("UNMOUNT", "Se ha desmontado correctamente la partición con ID: "+id)
				listaMount()
				return
			}
		}
	}
	Error("UNMOUNT", "No se encontró ninguna partición con el ID especificado.")

}

func GetMount(comando string, id string, p *string) Structs.Particion {
	if len(id) < 4 || id[len(id)-3:] != "576" {
		Error(comando, "El identificador no es válido.")
		return Structs.Particion{}
	}
	letra := id[len(id)-4]
	i := int(letra) - 'A'
	if i < 0 || i >= len(DiscMont) {
		Error(comando, "El identificador no es válido.")
		return Structs.Particion{}
	}
	for j := 0; j < 26; j++ {
		if DiscMont[i].Particiones[j].Estado == 1 {
			if DiscMont[i].Particiones[j].Letra == letra {

				path := ""
				for k := 0; k < len(DiscMont[i].Path); k++ {
					if DiscMont[i].Path[k] != 0 {
						path += string(DiscMont[i].Path[k])
					}
				}

				file, error := os.Open(strings.ReplaceAll(path, "\"", ""))
				if error != nil {
					Error(comando, "No se ha encontrado el disco")
					return Structs.Particion{}
				}
				disk := Structs.NewMBR()
				file.Seek(0, 0)

				data := leerBytes(file, int(unsafe.Sizeof(Structs.MBR{})))
				buffer := bytes.NewBuffer(data)
				err_ := binary.Read(buffer, binary.BigEndian, &disk)

				if err_ != nil {
					Error("FDSIK", "Error al leer el archivo")
					return Structs.Particion{}
				}
				file.Close()

				nombreParticion := ""
				for k := 0; k < len(DiscMont[i].Particiones[j].Nombre); k++ {
					if DiscMont[i].Particiones[j].Nombre[k] != 0 {
						nombreParticion += string(DiscMont[i].Particiones[j].Nombre[k])
					}
				}
				*p = path
				return *BuscarParticiones(disk, nombreParticion, path)
			}
		}
	}
	return Structs.Particion{}
}

func listaMount() {
	fmt.Println("\n<-------------------------- LISTADO DE MOUNTS -------------------------->")
	for i := 0; i < 99; i++ {
		for j := 0; j < 26; j++ {
			if DiscMont[i].Particiones[j].Estado == 1 {
				nombre := ""
				for k := 0; k < len(DiscMont[i].Particiones[j].Nombre); k++ {
					if DiscMont[i].Particiones[j].Nombre[k] != 0 {
						nombre += string(DiscMont[i].Particiones[j].Nombre[k])
					}
				}
				fmt.Println("\t id:" + string(alfabeto[j]) + "576" + ", Nombre: " + nombre)
			}
		}
	}
}
