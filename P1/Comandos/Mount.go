package Comandos

import (
	"P1/Structs"
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
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
	ID     string
	Letra  byte
	Estado byte
	Nombre [20]byte
}

var alfabeto = []byte{'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z'}
var cont = int(0)
var particionesDesmontadas = make(map[string]bool)

func ValidarDatosMOUNT(context []string) {
	name := ""
	driveletter := ""
	for i := 0; i < len(context); i++ {
		current := context[i]
		comando := strings.Split(current, "=")
		if Comparar(comando[0], "name") {
			name = comando[1]
		} else if Comparar(comando[0], "driveletter") {
			driveletter = comando[1]
		}
	}
	if driveletter == "" || name == "" {
		Error("MOUNT", "El comando MOUNT requiere parámetros obligatorios")
		return
	}
	mount(driveletter, name)
	listaMount()
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
		Error("UNMOUNT", "El comando UNMOUNT requiere el parametro -id para funcionar.")
		return
	}
	unmount(id)
}

func mount(p string, n string) {
	discoPath := "C:\\Users\\SuperUser\\Desktop\\Repositorio Local\\EJEMPLOS_MIA\\P1\\" + p + ".dsk"
	file, err := os.Open(discoPath)
	if err != nil {
		Error("Mount", "Error al abrir el archivo del disco")
		return
	}
	defer file.Close()

	disk := Structs.NewMBR()
	err = binary.Read(file, binary.BigEndian, &disk)
	if err != nil {
		Error("MOUNT", "Error al leer el MBR del disco.")
		return
	}

	file.Seek(0, 0)

	data := leerBytes(file, int(unsafe.Sizeof(Structs.MBR{})))
	buffer := bytes.NewBuffer(data)
	err_ := binary.Read(buffer, binary.BigEndian, &disk)
	if err_ != nil {
		Error("FDSIK", "Error al leer el archivo")
		return
	}
	file.Close()

	particion := BuscarParticiones(disk, n, p)
	if particion == nil {
		Error("MOUNT", "No se encontró la partición especificada en el disco.")
		return
	}

	if particion.Part_type == 'E' || particion.Part_type == 'L' {
		var nombre [16]byte
		copy(nombre[:], n)
		if particion.Part_name == nombre && particion.Part_type == 'E' {
			Error("MOUNT", "No se puede montar una partición extendida.")
			return
		} else {
			ebrs := GetLogicas(*particion, p)
			encontrada := false
			if len(ebrs) != 0 {
				for i := 0; i < len(ebrs); i++ {
					ebr := ebrs[i]
					nombreebr := ""
					for j := 0; j < len(ebr.Part_name); j++ {
						if ebr.Part_name[j] != 0 {
							nombreebr += string(ebr.Part_name[j])
						}
					}

					if Comparar(nombreebr, n) && ebr.Part_status == '1' {
						encontrada = true
						n = nombreebr
						break
					} else if nombreebr == n && ebr.Part_status == '0' {
						Error("MOUNT", "No se puede montar una partición Lógica eliminada.")
						return
					}
				}
				if !encontrada {
					Error("MOUNT", "No se encontró la partición Lógica.")
					return
				}
			}
		}
	}

	var id strings.Builder
	id.WriteByte(p[0]) // Agregar la primera letra del disco
	id.WriteString(strconv.Itoa(cont+1) + "76")
	cont++

	for j := 0; j < 26; j++ {
		var ruta [150]byte
		copy(ruta[:], p)
		if DiscMont[j].Path == ruta {
			for i := 0; i < 99; i++ {
				var nombre [20]byte
				copy(nombre[:], n)
				if DiscMont[j].Particiones[i].Nombre == nombre {
					Error("MOUNT", "Ya se ha montado la partición "+n)
					return
				}
				if DiscMont[j].Particiones[i].Estado == 0 {
					DiscMont[j].Particiones[i].Estado = 1
					DiscMont[j].Particiones[i].Letra = alfabeto[i]
					copy(DiscMont[j].Particiones[i].Nombre[:], n)
					DiscMont[j].Particiones[i].ID = id.String()
					Mensaje("MOUNT", "Se ha realizado correctamente el mount -id="+id.String())
					return
				}
			}
		}
	}

	for j := 0; j < 26; j++ {
		if DiscMont[j].Estado == 0 {
			DiscMont[j].Estado = 1
			copy(DiscMont[j].Path[:], p)
			for i := 0; i < 99; i++ {
				if DiscMont[j].Particiones[i].Estado == 0 {
					DiscMont[j].Particiones[i].Estado = 1
					DiscMont[j].Particiones[i].Letra = alfabeto[i]
					copy(DiscMont[j].Particiones[i].Nombre[:], n)
					DiscMont[j].Particiones[i].ID = id.String()
					Mensaje("MOUNT", "Se ha realizado correctamente el mount -id="+id.String())
					return
				}
			}
		}
	}
}

func listaMount() {
	fmt.Println("\n<-------------------------- LISTADO DE MOUNTS -------------------------->")
	for _, disco := range DiscMont {
		if disco.Estado == 1 {
			for _, particion := range disco.Particiones {
				if particion.Estado == 1 {
					fmt.Println("\t ID: " + particion.ID)
				}
			}
		}
	}
}

func unmount(id string) {
	if particionesDesmontadas[id] {
		Error("UNMOUNT", "La partición con ID "+id+" ya ha sido desmontada.")
		listaMount()
		return
	}
	for i := 0; i < 99; i++ {
		for j := 0; j < 26; j++ {
			if string(DiscMont[i].Particiones[j].ID) == id {
				DiscMont[i].Particiones[j].Estado = 0
				particionesDesmontadas[id] = true
				Mensaje("UNMOUNT", "La particion se ha desmontado de forma correcta")
				listaMount()
				return
			}
		}
	}
	Error("UNMOUNT", "No existe una particion con dicho ID")
}

func GetMount(comando string, id string, p *string) Structs.Particion {
	if !(id[2] == '5' && id[3] == '0') {
		Error(comando, "Formato de identificador incorrecto.")
		return Structs.Particion{}
	}

	letra := id[0]       // Obtener la letra del ID
	correlativo := "576" // Obtener el correlativo de partición del ID (sin incluir la letra)

	i, err := strconv.Atoi(correlativo)
	if err != nil || i < 1 || i > 99 {
		Error(comando, "Formato de identificador incorrecto.")
		return Structs.Particion{}
	}

	for j := 0; j < 26; j++ {
		if DiscMont[i-1].Particiones[j].Estado == 1 && DiscMont[i-1].Particiones[j].Letra == letra {
			path := ""
			for k := 0; k < len(DiscMont[i-1].Path); k++ {
				if DiscMont[i-1].Path[k] != 0 {
					path += string(DiscMont[i-1].Path[k])
				}
			}

			file, err := os.Open(strings.ReplaceAll(path, "\"", ""))
			if err != nil {
				Error(comando, "No se ha encontrado el disco")
				return Structs.Particion{}
			}
			defer file.Close()

			disk := Structs.NewMBR()
			file.Seek(0, 0)

			data := leerBytes(file, int(unsafe.Sizeof(Structs.MBR{})))
			buffer := bytes.NewBuffer(data)
			err = binary.Read(buffer, binary.BigEndian, &disk)

			if err != nil {
				Error("FDSIK", "Error al leer el archivo")
				return Structs.Particion{}
			}

			nombreParticion := ""
			for k := 0; k < len(DiscMont[i-1].Particiones[j].Nombre); k++ {
				if DiscMont[i-1].Particiones[j].Nombre[k] != 0 {
					nombreParticion += string(DiscMont[i-1].Particiones[j].Nombre[k])
				}
			}
			*p = path
			return *BuscarParticiones(disk, nombreParticion, path)
		}
	}

	Error(comando, "No se encontró la partición correspondiente en el disco montado.")
	return Structs.Particion{}
}

func idExists(id string) bool {
	for j := 0; j < len(DiscMont); j++ {
		for i := 0; i < len(DiscMont[j].Particiones); i++ {
			if DiscMont[j].Particiones[i].ID == id && DiscMont[j].Particiones[i].Estado == 1 {
				return true
			}
		}
	}
	return false
}
