package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("File upload request received")

	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Access-Control-Allow-Origin")

	// Check if the request method is OPTIONS (preflight request)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Parse the multipart form in the request
	err := r.ParseMultipartForm(10 << 20) // 10 MB limit
	if err != nil {
		http.Error(w, "Unable to parse multipart form", http.StatusBadRequest)
		log.Println("Error parsing multipart form:", err)
		return
	}

	// Get a reference to the fileHeaders
	files := r.MultipartForm.File["files"]

	// Iterate through the uploaded files
	for _, fileHeader := range files {
		// Open the uploaded file
		file, err := fileHeader.Open()
		if err != nil {
			http.Error(w, "Unable to open file", http.StatusInternalServerError)
			log.Println("Error opening file:", err)
			return
		}
		defer file.Close()

		// Create a new file in the server to save the uploaded file
		dst, err := os.Create(fileHeader.Filename)
		if err != nil {
			http.Error(w, "Unable to create the file", http.StatusInternalServerError)
			log.Println("Error creating file:", err)
			return
		}
		defer dst.Close()

		// Calculate file size
		fileSize := fileHeader.Size
		uploadedSize := int64(0) // Track uploaded bytes

		// Create a buffer to read chunks of the file
		buffer := make([]byte, 1024*1024) // 1MB buffer

		// Read chunks of the file and write to the destination file
		for {
			n, err := file.Read(buffer)
			if err != nil && err != io.EOF {
				http.Error(w, "Error reading file", http.StatusInternalServerError)
				log.Println("Error reading file:", err)
				return
			}
			if n == 0 {
				break
			}

			// Write the chunk to the destination file
			_, err = dst.Write(buffer[:n])
			if err != nil {
				http.Error(w, "Error writing to file", http.StatusInternalServerError)
				log.Println("Error writing to file:", err)
				return
			}

			// Update uploaded size
			uploadedSize += int64(n)

			// Calculate and log upload progress
			progress := float64(uploadedSize) / float64(fileSize) * 100
			log.Printf("File %s: %.2f%% uploaded\n", fileHeader.Filename, progress)
		}

		log.Println("File", fileHeader.Filename, "uploaded successfully.")
	}

	fmt.Println(len(files), "File(s) uploaded successfully.")
}

func main() {
	http.HandleFunc("/upload", uploadHandler)
	log.Println("Server started on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
