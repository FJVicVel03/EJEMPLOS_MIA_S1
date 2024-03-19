package Comandos

import (
	"P1/Structs"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type RepComando struct {
	Name string
	Path string
	ID   string
}

func ValidarDatosREPORT(context []string) {
	name := ""      // Inicializar variables fuera del bucle
	var path string // Declare the "path" variable
	path += "C:\\Users\\SuperUser\\Desktop\\Repositorio Local\\EJEMPLOS_MIA\\P1\\"
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
			path = path + tk[1]
		}
	}
	if id == "" {
		Error("REP", "El comando requiere el parámetro id obligatoriamente")
		return
	}
	if name == "" {
		Error("REP", "El comando requiere el parámetro name obligatoriamente")
	}
	if path == "" {
		Error("REP", "El comando requiere el parámetro path obligatoriamente")
	}
	generarReporte(name, path, id)
	// Ahora name y ruta están siendo utilizados en el alcance de la función
}

func generarReporte(name, path, id string) {
	mbr, _, _ := LeerDatosParticionMontada(id)
	if mbr == nil {
		Error("REP", "Error al leer el MBR del disco montado")
		return
	}

	// Verificar y crear el directorio si no existe
	if err := verificarDirectorio(path); err != nil {
		Error("REP", "Error al verificar o crear el directorio: "+err.Error())
		return
	}

	// Obtener el nombre del archivo del path proporcionado
	dir, fileNameWithExt := filepath.Split(path)
	fileName := strings.TrimSuffix(fileNameWithExt, filepath.Ext(fileNameWithExt))

	// Crear el archivo DOT para el reporte del MBR
	dotContent := "digraph MBR_Report {\n"
	dotContent += "\tlabelloc=top\n"
	dotContent += "\trankdir=TB\n"
	dotContent += "\tnode [shape=plaintext]\n"
	dotContent += "\tedge [style=invis]\n"
	dotContent += "\ttable [\n"
	dotContent += "\t\tlabel=<<table border=\"1\" cellborder=\"1\" cellspacing=\"0\">\n"
	dotContent += "\t\t\t<tr><td colspan=\"2\"> Reporte MBR </td></tr>\n"
	dotContent += fmt.Sprintf("\t\t\t<tr><td>tamano</td><td>%d</td></tr>\n", mbr.Mbr_Tamano)
	dotContent += fmt.Sprintf("\t\t\t<tr><td>fecha_creacion</td><td>%s</td></tr>\n", string(mbr.Mbr_fecha_creacion[:]))
	dotContent += fmt.Sprintf("\t\t\t<tr><td>disk_signature</td><td>%d</td></tr>\n", mbr.Mbr_disk_signature)

	// Agregar la información de las particiones
	dotContent += "\t\t\t<tr><td colspan=\"2\"> Particiones </td></tr>\n"
	for _, particion := range mbr.Particiones {
		// Limpiar el nombre de la partición de caracteres nulos
		particionNombre := strings.TrimRight(string(particion.Part_name[:]), "\x00")
		// Construir la tabla de datos de la partición
		particionTable := fmt.Sprintf("\t\t\t<tr><td colspan=\"2\"> Particion </td></tr>\n"+
			"\t\t\t<tr><td> Status </td><td>%c</td></tr>\n"+
			"\t\t\t<tr><td> Tipo </td><td>%c</td></tr>\n"+
			"\t\t\t<tr><td> Fit </td><td>%c</td></tr>\n"+
			"\t\t\t<tr><td> Start </td><td>%d</td></tr>\n"+
			"\t\t\t<tr><td> Size </td><td>%d</td></tr>\n"+
			"\t\t\t<tr><td> Nombre </td><td>%s</td></tr>\n", particion.Part_status, particion.Part_type, particion.Part_fit, particion.Part_start, particion.Part_size, particionNombre)
		// Agregar la tabla de partición al contenido DOT
		dotContent += particionTable
	}

	dotContent += "\t\t</table>>\n"
	dotContent += "\t]\n"
	dotContent += "}\n"

	dotFilePath := filepath.Join(dir, fileName+".dot")
	err := guardarArchivo(dotFilePath, []byte(dotContent))
	if err != nil {
		Error("REP", "Error al guardar el archivo DOT del reporte del MBR: "+err.Error())
		return
	}

	// Generar la imagen JPG utilizando Graphviz
	jpgFilePath := filepath.Join(dir, fileName)
	err = generarImagenDOT(dotFilePath, jpgFilePath)
	if err != nil {
		Error("REP", "Error al generar la imagen JPG del reporte del MBR: "+err.Error())
		return
	}

	// Reporte generado con éxito
	Mensaje("REP", "Reporte generado con éxito y guardado en: "+jpgFilePath)
}

// Estructura para almacenar información relevante de una partición
type InformacionParticion struct {
	Nombre         string
	Tipo           string
	Tamano         int64
	PosicionInicio int64
	// Puedes agregar más campos según sea necesario
}

func guardarArchivo(filePath string, data []byte) error {
	return ioutil.WriteFile(filePath, data, 0644)
}

func generarImagenDOT(dotFilePath, jpgFilePath string) error {
	cmd := exec.Command("dot", "-Tjpg", "-o", jpgFilePath, dotFilePath)
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func LeerDatosParticionMontada(id string) (*Structs.MBR, *Structs.Particion, *Structs.EBR) {
	// Obtener el disco montado usando su ID
	var discoMontado DiscoMontado
	encontrado := false
	for _, disco := range DiscMont {
		for _, particion := range disco.Particiones {
			if particion.ID == id {
				discoMontado = disco
				encontrado = true
				break
			}
		}
		if encontrado {
			break
		}
	}

	if !encontrado {
		fmt.Println("No se encontró la partición montada con el ID especificado:", id)
		return nil, nil, nil
	}

	// Leer el MBR del disco montado
	// Aquí necesitas proporcionar la ruta correcta del disco, actualmente solo estás pasando la primera letra del ID
	path := fmt.Sprintf("C:\\Users\\SuperUser\\Desktop\\Repositorio Local\\EJEMPLOS_MIA\\P1\\%s.dsk", id[:1])
	mbr := leerDisco(path)
	if mbr == nil {
		fmt.Println("Error al leer el MBR del disco montado:", path)
		return nil, nil, nil
	}

	// Obtener la partición correspondiente del MBR
	var particion *Structs.Particion
	for i := range mbr.Particiones {
		if string(mbr.Particiones[i].Part_name[:]) == string(discoMontado.Particiones[0].Nombre[:]) {
			particion = &mbr.Particiones[i]
			break
		}
	}

	// Verificar si la partición es extendida para leer los EBRs
	if particion != nil && particion.Part_type == 'E' {
		ebrs := GetLogicas(*particion, path)
		if len(ebrs) == 0 {
			fmt.Println("No se encontraron EBRs para la partición extendida:", string(particion.Part_name[:]))
			return mbr, particion, nil
		}

		// Encontrar el EBR correspondiente usando el nombre de la partición montada
		var ebr *Structs.EBR
		for _, e := range ebrs {
			if string(e.Part_name[:]) == string(discoMontado.Particiones[0].Nombre[:]) {
				ebr = &e
				break
			}
		}
		if ebr == nil {
			fmt.Println("No se encontró el EBR correspondiente para la partición lógica:", string(discoMontado.Particiones[0].Nombre[:]))
		}

		return mbr, particion, ebr
	}

	return mbr, particion, nil
}

func verificarDirectorio(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}
