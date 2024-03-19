package Comandos

import (
	"P1/Structs"
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unsafe"
)

type Transition struct {
	partition int
	start     int
	end       int
	before    int
	after     int
}

var startValue int

func ValidarDatosFDISK(tokens []string) {
	size := ""
	unit := "k"
	driveletter := ""
	tipo := "P"
	fit := "WF"
	name := ""
	add := ""
	borrar := ""
	for i := 0; i < len(tokens); i++ {
		token := tokens[i]
		tk := strings.Split(token, "=")
		if Comparar(tk[0], "size") {
			size = tk[1]
		} else if Comparar(tk[0], "unit") {
			unit = tk[1]
		} else if Comparar(tk[0], "driveletter") {
			driveletter = tk[1]
		} else if Comparar(tk[0], "type") {
			tipo = tk[1]
		} else if Comparar(tk[0], "fit") {
			fit = tk[1]
		} else if Comparar(tk[0], "name") {
			name = tk[1]
		} else if Comparar(tk[0], "add") {
			add = tk[1]
		} else if Comparar(tk[0], "delete") {
			borrar = tk[1]
		} else {
			Error("FDISK", "El comando FDISK no acepta este comando"+tk[0])
			return
		}
	}

	if add != "" {
		addSpace(driveletter, name, add, unit)
		return
	}

	if borrar != "" {
		borrarparticion(driveletter, name, borrar)
		return
	}

	if driveletter == "" || name == "" {
		Error("FDISK", "El comando FDISK necesita parametros obligatorios")
		return
	} else {
		generarParticion(size, unit, driveletter, tipo, fit, name)
	}
}

func generarParticion(s string, u string, p string, t string, f string, n string) {
	p += ".dsk"
	startValue = 0
	i, error_ := strconv.Atoi(s)
	if error_ != nil {
		Error("FDISK", "Size debe ser un número entero")
		return
	}
	if i <= 0 {
		Error("FDISK", "Size debe ser mayor que 0")
		return
	}
	if Comparar(u, "b") || Comparar(u, "k") || Comparar(u, "m") {
		if Comparar(u, "k") {
			i = i * 1024
		} else if Comparar(u, "m") {
			i = i * 1024 * 1024
		}
	} else {
		Error("FDISK", "Unit no contiene los valores esperados.")
		return
	}
	if !(Comparar(t, "p") || Comparar(t, "e") || Comparar(t, "l")) {
		Error("FDISK", "Type no contiene los valores esperados.")
		return
	}
	if !(Comparar(f, "bf") || Comparar(f, "ff") || Comparar(f, "wf")) {
		Error("FDISK", "Fit no contiene los valores esperados.")
		return
	}

	discoPath := "C:\\Users\\SuperUser\\Desktop\\Repositorio Local\\EJEMPLOS_MIA\\P1"

	path := filepath.Join(discoPath, p)

	mbr := leerDisco(path)
	if mbr == nil {
		return
	}

	if int64(i) > mbr.Mbr_Tamano {
		Error("FDISK", "El tamaño de la partición es mayor que el tamaño del disco duro.")
		return
	}

	particiones := GetParticiones(*mbr)
	var between []Transition

	usado := 0
	ext := 0
	c := 0
	base := int(unsafe.Sizeof(Structs.MBR{}))
	extended := Structs.NewParticion()

	for j := 0; j < len(particiones); j++ {
		prttn := particiones[j]
		if prttn.Part_status == '1' {
			var trn Transition
			trn.partition = c
			trn.start = int(prttn.Part_start)
			trn.end = int(prttn.Part_start + prttn.Part_size)
			trn.before = trn.start - base
			base = trn.end
			if usado != 0 {
				between[usado-1].after = trn.start - (between[usado-1].end)
			}
			between = append(between, trn)
			usado++

			if prttn.Part_type == "e"[0] || prttn.Part_type == "E"[0] {
				ext++
				extended = prttn
			}
		}
		if usado == 4 && !Comparar(t, "l") {
			Error("FDISK", "Limite de particiones alcanzado")
			return
		} else if ext == 1 && Comparar(t, "e") {
			Error("FDISK", "Solo se puede crear una partición extendida")
			return
		}
		c++
	}
	if ext == 0 && Comparar(t, "l") {
		Error("FDISK", "Aún no se han creado particiones extendidas, no se puede agregar una lógica.")
		return
	}
	if usado != 0 {
		between[len(between)-1].after = int(mbr.Mbr_Tamano) - between[len(between)-1].end
	}
	regresa := BuscarParticiones(*mbr, n, p)
	if regresa != nil {
		Error("FDISK", "El nombre: "+n+", ya está en uso.")
		return
	}
	temporal := Structs.NewParticion()
	temporal.Part_status = '1'
	temporal.Part_size = int64(i)
	temporal.Part_type = strings.ToUpper(t)[0]
	temporal.Part_fit = strings.ToUpper(f)[0]
	copy(temporal.Part_name[:], n)

	if Comparar(t, "l") {
		Logica(temporal, extended, p)
		return
	}
	mbr = ajustar(*mbr, temporal, between, particiones, usado)
	if mbr == nil {
		return
	}
	file, err := os.OpenFile(strings.ReplaceAll(filepath.Join(discoPath, p), "\"", ""), os.O_WRONLY, os.ModeAppend)
	if err != nil {
		Error("FDISK", "Error al abrir el archivo 2: "+err.Error())
		return
	}
	defer file.Close()
	file.Seek(0, 0)
	var binario2 bytes.Buffer
	binary.Write(&binario2, binary.BigEndian, mbr)
	EscribirBytes(file, binario2.Bytes())
	if Comparar(t, "E") {
		ebr := Structs.NewEBR()
		ebr.Part_status = '0'
		ebr.Part_start = int64(startValue)
		ebr.Part_size = 0
		ebr.Part_next = -1

		file.Seek(int64(startValue), 0) //5200
		var binario3 bytes.Buffer
		binary.Write(&binario3, binary.BigEndian, ebr)
		EscribirBytes(file, binario3.Bytes())
		Mensaje("FDISK", "Partición Extendida: "+n+", creada correctamente.")
		return
	}
	file.Close()
	Mensaje("FDISK", "Partición Primaria: "+n+", creada correctamente.")
}

func GetParticiones(disco Structs.MBR) []Structs.Particion {
	var v []Structs.Particion
	v = append(v, disco.Particiones[0])
	v = append(v, disco.Particiones[1])
	v = append(v, disco.Particiones[2])
	v = append(v, disco.Particiones[3])
	return v
}

func BuscarParticiones(mbr Structs.MBR, name string, path string) *Structs.Particion {
	var particiones [4]Structs.Particion
	particiones[0] = mbr.Particiones[0]
	particiones[1] = mbr.Particiones[1]
	particiones[2] = mbr.Particiones[2]
	particiones[3] = mbr.Particiones[3]

	ext := false
	extended := Structs.NewParticion()
	for i := 0; i < len(particiones); i++ {
		particion := particiones[i]
		if particion.Part_status == "1"[0] {
			nombre := ""
			for j := 0; j < len(particion.Part_name); j++ {
				if particion.Part_name[j] != 0 {
					nombre += string(particion.Part_name[j])
				}
			}
			if Comparar(nombre, name) {
				return &particion
			} else if particion.Part_type == "E"[0] || particion.Part_type == "e"[0] {
				ext = true
				extended = particion
			}
		}
	}

	if ext {
		ebrs := GetLogicas(extended, path)
		for i := 0; i < len(ebrs); i++ {
			ebr := ebrs[i]
			if ebr.Part_status == '1' {
				nombre := ""
				for j := 0; j < len(ebr.Part_name); j++ {
					if ebr.Part_name[j] != 0 {
						nombre += string(ebr.Part_name[j])
					}
				}
				if Comparar(nombre, name) {
					tmp := Structs.NewParticion()
					tmp.Part_status = '1'
					tmp.Part_type = 'L'
					tmp.Part_fit = ebr.Part_fit
					tmp.Part_start = ebr.Part_start
					tmp.Part_size = ebr.Part_size
					copy(tmp.Part_name[:], ebr.Part_name[:])
					return &tmp
				}
			}
		}
	}
	return nil
}

func GetLogicas(particion Structs.Particion, path string) []Structs.EBR {
	var ebrs []Structs.EBR
	discoPath := "C:\\Users\\SuperUser\\Desktop\\Repositorio Local\\EJEMPLOS_MIA\\P1" + path + ".dsk"
	file, err := os.Open(discoPath)
	if err != nil {
		Error("FDISK", "Error al abrir el archivo del disco")
		return nil
	}
	defer file.Close()
	file.Seek(0, 0)
	tmp := Structs.NewEBR()
	file.Seek(particion.Part_start, 0)

	data := leerBytes(file, int(unsafe.Sizeof(Structs.EBR{})))
	buffer := bytes.NewBuffer(data)
	err_ := binary.Read(buffer, binary.BigEndian, &tmp)
	if err_ != nil {
		Error("FDSIK", "Error al leer el archivo")
		return nil
	}
	for {
		if int(tmp.Part_next) != -1 && int(tmp.Part_status) != 0 {
			ebrs = append(ebrs, tmp)
			file.Seek(tmp.Part_next, 0)

			data = leerBytes(file, int(unsafe.Sizeof(Structs.EBR{})))
			buffer = bytes.NewBuffer(data)
			err_ = binary.Read(buffer, binary.BigEndian, &tmp)
			if err_ != nil {
				Error("FDSIK", "Error al leer el archivo")
				return nil
			}
		} else {
			file.Close()
			break
		}
	}

	return ebrs
}

func Logica(particion Structs.Particion, ep Structs.Particion, driveletter string) {
	discoPath := "C:\\Users\\SuperUser\\Desktop\\Repositorio Local\\EJEMPLOS_MIA\\P1" + driveletter + ".dsk"
	logic := Structs.NewEBR()
	logic.Part_status = '1'
	logic.Part_fit = particion.Part_fit
	logic.Part_size = particion.Part_size
	logic.Part_next = -1
	copy(logic.Part_name[:], particion.Part_name[:])

	file, err := os.Open(discoPath)
	if err != nil {
		Error("FDISK", "Error al abrir el archivo del disco.")
		return
	}
	defer file.Close()

	tmp := Structs.NewEBR()
	tmp.Part_status = 0
	tmp.Part_size = 0
	tmp.Part_next = -1
	file.Seek(ep.Part_start, 0) //0

	data := leerBytes(file, int(unsafe.Sizeof(Structs.EBR{})))
	buffer := bytes.NewBuffer(data)
	err = binary.Read(buffer, binary.BigEndian, &tmp)

	if err != nil {
		Error("FDSIK", "Error al leer el archivo")
		return
	}
	if err != nil {
		Error("FDISK", "Error al abrir el archivo del disco.")
		return
	}
	var size int64 = 0
	file.Close()
	for {
		size += int64(unsafe.Sizeof(Structs.EBR{})) + tmp.Part_size
		if (tmp.Part_size == 0 && tmp.Part_next == -1) || (tmp.Part_size == 0 && tmp.Part_next == 0) {
			file, err = os.OpenFile(discoPath, os.O_WRONLY, os.ModeAppend)
			logic.Part_start = tmp.Part_start
			logic.Part_next = logic.Part_start + logic.Part_size + int64(unsafe.Sizeof(Structs.EBR{}))
			if (ep.Part_size - size) <= logic.Part_size {
				Error("FDISK", "No queda más espacio para crear más particiones lógicas")
				return
			}
			file.Seek(logic.Part_start, 0)

			var binario2 bytes.Buffer
			binary.Write(&binario2, binary.BigEndian, logic)
			EscribirBytes(file, binario2.Bytes())
			nombre := ""
			for j := 0; j < len(particion.Part_name); j++ {
				nombre += string(particion.Part_name[j])
			}
			file.Seek(logic.Part_next, 0)
			addLogic := Structs.NewEBR()
			addLogic.Part_status = '0'
			addLogic.Part_next = -1
			addLogic.Part_start = logic.Part_next

			file.Seek(addLogic.Part_start, 0)

			var binarioLogico bytes.Buffer
			binary.Write(&binarioLogico, binary.BigEndian, addLogic)
			EscribirBytes(file, binarioLogico.Bytes())

			Mensaje("FDISK", "Partición Lógica: "+nombre+", creada correctamente.")
			file.Close()
			return
		}
		file, err = os.Open(discoPath)
		if err != nil {
			Error("FDISK", "Error al abrir el archivo del disco.")
			return
		}
		file.Seek(tmp.Part_next, 0)
		data = leerBytes(file, int(unsafe.Sizeof(Structs.EBR{})))
		buffer = bytes.NewBuffer(data)
		err = binary.Read(buffer, binary.BigEndian, &tmp)

		if err != nil {
			Error("FDSIK", "Error al leer el archivo")
			return
		}
	}
}

func ajustar(mbr Structs.MBR, p Structs.Particion, t []Transition, ps []Structs.Particion, u int) *Structs.MBR {
	if u == 0 {
		p.Part_start = int64(unsafe.Sizeof(mbr))
		startValue = int(p.Part_start)
		mbr.Particiones[0] = p
		return &mbr
	} else {
		var usar Transition
		c := 0
		for i := 0; i < len(t); i++ {
			tr := t[i]
			if c == 0 {
				usar = tr
				c++
				continue
			}

			if Comparar(string(mbr.Dsk_fit[0]), "F") {
				if int64(usar.before) >= p.Part_size || int64(usar.after) >= p.Part_size {
					break
				}
				usar = tr
			} else if Comparar(string(mbr.Dsk_fit[0]), "B") {
				if int64(tr.before) >= p.Part_size || int64(usar.after) < p.Part_size {
					usar = tr
				} else {
					if int64(tr.before) >= p.Part_size || int64(tr.after) >= p.Part_size {
						b1 := usar.before - int(p.Part_size)
						a1 := usar.after - int(p.Part_size)
						b2 := tr.before - int(p.Part_size)
						a2 := tr.after - int(p.Part_size)

						if (b1 < b2 && b1 < a2) || (a1 < b2 && a1 < a2) {
							c++
							continue
						}
						usar = tr
					}
				}
			} else if Comparar(string(mbr.Dsk_fit[0]), "W") {
				if int64(usar.before) >= p.Part_size || int64(usar.after) < p.Part_size {
					usar = tr
				} else {
					if int64(tr.before) >= p.Part_size || int64(tr.after) >= p.Part_size {
						b1 := usar.before - int(p.Part_size)
						a1 := usar.after - int(p.Part_size)
						b2 := tr.before - int(p.Part_size)
						a2 := tr.after - int(p.Part_size)

						if (b1 > b2 && b1 > a2) || (a1 > b2 && a1 > a2) {
							c++
							continue
						}
						usar = tr
					}
				}
			}
			c++
		}
		if usar.before >= int(p.Part_size) || usar.after >= int(p.Part_size) {
			if Comparar(string(mbr.Dsk_fit[0]), "F") {
				if usar.before >= int(p.Part_size) {
					p.Part_start = int64(usar.start - usar.before)
					startValue = int(p.Part_start)
				} else {
					p.Part_start = int64(usar.end)
					startValue = int(p.Part_start)
				}
			} else if Comparar(string(mbr.Dsk_fit[0]), "B") {
				b1 := usar.before - int(p.Part_size)
				a1 := usar.after - int(p.Part_size)

				if (usar.before >= int(p.Part_size) && b1 < a1) || usar.after < int(p.Part_start) {
					p.Part_start = int64(usar.start - usar.before)
					startValue = int(p.Part_start)
				} else {
					p.Part_start = int64(usar.end)
					startValue = int(p.Part_start)
				}
			} else if Comparar(string(mbr.Dsk_fit[0]), "W") {
				b1 := usar.before - int(p.Part_size)
				a1 := usar.after - int(p.Part_size)

				if (usar.before >= int(p.Part_size) && b1 > a1) || usar.after < int(p.Part_start) {
					p.Part_start = int64(usar.start - usar.before)
					startValue = int(p.Part_start)
				} else {
					p.Part_start = int64(usar.end)
					startValue = int(p.Part_start)
				}
			}
			var partitions [4]Structs.Particion
			for i := 0; i < len(ps); i++ {
				partitions[i] = ps[i]
			}

			for i := 0; i < len(partitions); i++ {
				partition := partitions[i]
				if partition.Part_status != '1' {
					partitions[i] = p
					break
				}
			}
			mbr.Particiones[0] = partitions[0]
			mbr.Particiones[1] = partitions[1]
			mbr.Particiones[2] = partitions[2]
			mbr.Particiones[3] = partitions[3]
			return &mbr
		} else {
			Error("FDISK", "No hay espacio suficiente.")
			return nil
		}
	}
}

func addSpace(letra string, nombre string, espacio string, unidad string) {
	disco := leerDisco(letra)
	discoPath := "C:\\Users\\SuperUser\\Desktop\\Repositorio Local\\EJEMPLOS_MIA\\P1" + letra + ".dsk"
	if disco == nil {
		Error("FDISK", "El disco seleccionado no existe.")
		return
	}
	particion := BuscarParticiones(*disco, nombre, letra)
	if particion == nil {
		Error("FDISK", "La particion seleccionada no existe.")
		return
	}
	nuevo_espacio, err := strconv.Atoi(espacio)
	if err != nil {
		Error("FDISK", "Ocurrio un error al agregar el tamaño.")
		return
	}
	if Comparar(unidad, "b") || Comparar(unidad, "k") || Comparar(unidad, "m") {
		if Comparar(unidad, "k") {
			nuevo_espacio = nuevo_espacio * 1024
		} else if Comparar(unidad, "m") {
			nuevo_espacio = nuevo_espacio * 1024 * 1024
		}
	}
	if int64(nuevo_espacio) > disco.Mbr_Tamano {
		Error("FDISK", "El espacio que se quiere agregar en la particion es mayor al tamaño del disco duro. ")
		return
	}

	agregar_data := particion.Part_size + int64(nuevo_espacio)

	if agregar_data < 0 {
		Error("FDISK", "La particion no puede ser negativa.")
		return
	}

	particion.Part_size += int64(nuevo_espacio)

	file, err := os.OpenFile(discoPath, os.O_WRONLY, os.ModeAppend)
	if err != nil {
		Error("FDISK", "Error al abrir el archivo 3: "+err.Error())
		return
	}
	defer file.Close()
	file.Seek(0, 0)
	var binario2 bytes.Buffer
	binary.Write(&binario2, binary.BigEndian, disco)
	EscribirBytes(file, binario2.Bytes())
	Mensaje("FDISK", "El nuevo tamaño de la particion: "+string(particion.Part_name[:])+" ahora es:"+strconv.FormatInt(particion.Part_size, 10))
	file.Close()
}

func borrarparticion(letra string, nombre string, borrar string) {
	disco := leerDisco(letra)
	discoPath := "C:\\Users\\SuperUser\\Desktop\\Repositorio Local\\EJEMPLOS_MIA\\P1" + letra + ".dsk"
	if disco == nil {
		Error("FDISK", "El disco seleccionado no existe.")
		return
	}
	particion := BuscarParticiones(*disco, nombre, letra)
	if particion == nil {
		Error("FDISK", "La particion seleccionada no existe.")
		return
	}
	if borrar == "full" {
		particion.Part_status = 0
		file, err := os.OpenFile(discoPath, os.O_WRONLY, os.ModeAppend)
		if err != nil {
			Error("FDISK", "Error al abrir el archivo 1: "+err.Error())
			return
		}
		defer file.Close()
		file.Seek(particion.Part_start, 0)
		nullBytes := make([]byte, particion.Part_size)
		for i := range nullBytes {
			nullBytes[i] = 0
		}

		_, err = file.Write(nullBytes)
		if err != nil {
			Error("FDISK", "Error al escribir los bytes en el disco")
			return
		}

		Mensaje("FDISK", "Espacio marcado como vacío y rellenado con \\0 en la partición "+nombre)
		return
	} else {
		Error("FDISK", "Unicamente se puede usar la palabra \"full\"")
		return
	}
}
