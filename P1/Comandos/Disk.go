package Comandos

import (
	"P1/Structs"
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"math/big"
	"os"
	"strconv"
	"strings"
	"time"
)

func ValidarDatosMKDISK(tokens []string) {
	size := ""
	fit := ""
	unit := ""
	error_ := false
	for i := 0; i < len(tokens); i++ {
		token := tokens[i]
		tk := strings.Split(token, "=")
		if Comparar(tk[0], "fit") {
			if fit == "" {
				fit = tk[1]
			} else {
				Error("MKDISK", "parametro f repetido en el comando: "+tk[0])
				return
			}
		} else if Comparar(tk[0], "size") {
			if size == "" {
				size = tk[1]
			} else {
				Error("MKDISK", "parametro SIZE repetido en el comando: "+tk[0])
				return
			}
		} else if Comparar(tk[0], "unit") {
			if unit == "" {
				unit = tk[1]
			} else {
				Error("MKDISK", "parametro U repetido en el comando: "+tk[0])
				return
			}
		} else {
			Error("MKDISK", "no se esperaba el parametro "+tk[0])
			error_ = true
			return
		}
	}
	if error_ {
		return
	}
	if size == "" {
		Error("MKDISK", "se requiere parametro Size para este comando")
		return
	}
	if fit == "" {
		fit = "FF"
	}
	if !Comparar(fit, "BF") && !Comparar(fit, "FF") && !Comparar(fit, "WF") {
		Error("MKDISK", "valores en parametro fit no esperados")
		return
	}
	if unit == "" {
		unit = "m"
	}
	if !Comparar(unit, "k") && !Comparar(unit, "m") {
		Error("MKDISK", "valores en parametro unit no esperados")
		return
	}

	// Llamar a makeFile sin proporcionar el path
	makeFile(size, fit, unit)
}

func makeFile(s string, f string, u string) {
	var disco = Structs.NewMBR()
	size, err := strconv.Atoi(s)
	if err != nil {
		Error("MKDISK", "Size debe ser un número entero")
		return
	}
	if size <= 0 {
		Error("MKDISK", "Size debe ser mayor a 0")
		return
	}
	if Comparar(u, "M") {
		size = 1024 * 1024 * size
	} else if Comparar(u, "k") {
		size = 1024 * size
	}
	f = string(f[0])

	disco.Mbr_Tamano = int64(size)
	fecha := time.Now().String()
	copy(disco.Mbr_fecha_creacion[:], fecha)
	aleatorio, _ := rand.Int(rand.Reader, big.NewInt(999999999))
	entero, _ := strconv.Atoi(aleatorio.String())
	disco.Mbr_disk_signature = int64(entero)
	copy(disco.Dsk_fit[:], string(f[0]))
	disco.Particiones[0] = Structs.NewParticion()
	disco.Particiones[1] = Structs.NewParticion()
	disco.Particiones[2] = Structs.NewParticion()
	disco.Particiones[3] = Structs.NewParticion()

	// Generar el nombre del disco automáticamente (de la A a la Z)

	for letra := 'A'; letra <= 'Z'; letra++ {
		path := string(letra) + ".dsk"
		if !ArchivoExiste(path) {
			carpeta := "C:\\Users\\SuperUser\\Desktop\\Repositorio Local\\EJEMPLOS_MIA\\P1"
			direccion := strings.Split(path, "/")
			for i := 0; i < len(direccion)-1; i++ {
				carpeta += "/" + direccion[i]
				if _, err_ := os.Stat(carpeta); os.IsNotExist(err_) {
					os.Mkdir(carpeta, 0777)
				}
			}

			// Crear el archivo .dsk
			file, err := os.Create(path)
			if err != nil {
				Error("MKDISK", "No se pudo crear el disco: "+err.Error())
				return
			}
			defer file.Close() // Asegurar que el archivo se cierre al salir de la función
			// Escribir la información del MBR en el archivo
			var binario3 bytes.Buffer
			binary.Write(&binario3, binary.BigEndian, disco)
			EscribirBytes(file, binario3.Bytes())

			Mensaje("MKDISK", "¡Disco \""+path+"\" creado correctamente!")
			return
		}
	}
	Error("MKDISK", "No se pudo crear el disco: no hay nombres de disco disponibles.")
}

func RMDISK(tokens []string) {
	if len(tokens) != 1 {
		Error("RMDISK", "Se espera un parámetro de driveletter.")
		return
	}

	driveLetter := ""
	for i := 0; i < len(tokens); i++ {
		token := tokens[i]
		tk := strings.Split(token, "=")
		if Comparar(tk[0], "driveletter") {
			if driveLetter == "" {
				driveLetter = tk[1]
			} else {
				Error("RMDISK", "Parámetro driveletter repetido en el comando: "+tk[0])
				return
			}
		} else {
			Error("RMDISK", "Parámetro no esperado: "+tk[0])
			return
		}
	}

	if driveLetter == "" {
		Error("RMDISK", "Se requiere el parámetro driveletter para este comando.")
		return
	}

	diskPath := driveLetter + ".dsk"

	if !ArchivoExiste(diskPath) {
		Error("RMDISK", "No se encontró el disco en la ruta indicada.")
		return
	}

	if Confirmar("¿Desea eliminar el disco: " + diskPath + " ?") {
		err := os.Remove(diskPath)
		if err != nil {
			Error("RMDISK", "Error al intentar eliminar el archivo: "+err.Error())
			return
		}
		Mensaje("RMDISK", "El disco "+diskPath+" ha sido eliminado exitosamente.")
		return
	} else {
		Mensaje("RMDISK", "Eliminación del disco "+diskPath+" cancelada.")
		return
	}
}
