package main

import (
	"encoding/base64"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Estructura para pasar las imágenes y el nombre del host al template HTML
type PageData struct {
	Images   []string
	Hostname string
}

// Variable global para almacenar el directorio de imágenes
var imageDirectory string
var once sync.Once

// Procesar las imágenes en base64
func encodeImagesToBase64(imagePaths []string) ([]string, error) {
	var base64Images []string
	for _, imagePath := range imagePaths {
		imageBytes, err := ioutil.ReadFile(imagePath)
		if err != nil {
			return nil, err
		}
		encoded := base64.StdEncoding.EncodeToString(imageBytes)
		base64Images = append(base64Images, encoded)
	}
	return base64Images, nil
}

// Seleccionar 4 imágenes aleatorias de la carpeta
func selectRandomImages(folderPath string) ([]string, error) {
	var images []string
	err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && (strings.HasSuffix(strings.ToLower(path), ".jpg") || strings.HasSuffix(strings.ToLower(path), ".jpeg") || strings.HasSuffix(strings.ToLower(path), ".png")) {
			images = append(images, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Seleccionar 4 imágenes aleatorias
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(images), func(i, j int) { images[i], images[j] = images[j], images[i] })
	if len(images) > 4 {
		images = images[:4]
	}
	return images, nil
}

// Manejar la página principal
func handler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("index.html"))

	// Obtener el nombre del host
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "Desconocido"
	}

	// Usar el directorio de imágenes almacenado globalmente
	selectedImages, err := selectRandomImages(imageDirectory)
	if err != nil {
		http.Error(w, "Error al seleccionar imágenes", http.StatusInternalServerError)
		return
	}

	// Convertir imágenes a base64
	base64Images, err := encodeImagesToBase64(selectedImages)
	if err != nil {
		http.Error(w, "Error al procesar imágenes", http.StatusInternalServerError)
		return
	}

	// Pasar los datos (imágenes y nombre del host) al template
	data := PageData{
		Images:   base64Images,
		Hostname: hostname,
	}

	// Renderizar la plantilla HTML
	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Error al renderizar HTML", http.StatusInternalServerError)
	}
}

func main() {
	// Usar sync.Once para asegurarnos que el directorio se pida solo una vez
	once.Do(func() {
		// Solicitar ruta de imágenes al usuario
		fmt.Println("Ingresa la ruta de la carpeta con imágenes (jpg, jpeg, png):")
		fmt.Scanln(&imageDirectory)
		//
		imageDirectory = "home/asus/Walpaper"

		// Validar que el directorio exista
		if _, err := os.Stat(imageDirectory); os.IsNotExist(err) {
			log.Fatalf("La carpeta especificada no existe: %s", imageDirectory)
		}
	})

	// Solicitar puerto de ejecución al usuario
	//fmt.Println("Ingresa el puerto de ejecución:")
	var port string
	//fmt.Scanln(&port)
	port = "8000"
	// Manejar archivos estáticos (imágenes, íconos, etc.)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Manejar la página principal
	http.HandleFunc("/", handler)

	fmt.Printf("Servidor iniciado en http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
