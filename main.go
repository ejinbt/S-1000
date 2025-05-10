// SCRAPPER-1000
// author : EJIN

package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Message struct {
	Author Author `json:"author"`
}

type Author struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Nickname string `json:"nickname"`
	Roles    []Role `json:"roles"`
}

type Role struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type UserRoleEntry struct {
	UserID      string
	Username    string
	DisplayName string
	RoleID      string
	RoleName    string
}

func processJSONFile(filePath string, uniqueEntries map[string]UserRoleEntry, mu *sync.Mutex, wg *sync.WaitGroup) {
	defer wg.Done()

	log.Printf("Processing file: %s\n", filePath)
	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("Error opening file %s: %v\n", filePath, err)
		return
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	decoder := json.NewDecoder(reader)

	_, err = decoder.Token()
	if err != nil {
		log.Printf("Error reading opening brace from %s: %v\n", filePath, err)
		return
	}

	foundMessages := false
	for decoder.More() {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Error reading token key from %s: %v\n", filePath, err)
			return
		}

		key, ok := token.(string)
		if !ok {
			var dummy interface{}
			if err := decoder.Decode(&dummy); err != nil && err != io.EOF {
				log.Printf("Error skipping unexpected value in %s: %v\n", filePath, err)
				return
			}
			continue
		}

		if key == "messages" {
			foundMessages = true
			_, err = decoder.Token()
			if err != nil {
				log.Printf("Error reading opening bracket for messages array in %s: %v\n", filePath, err)
				return
			}

			for decoder.More() {
				var msg Message
				if err := decoder.Decode(&msg); err != nil {
					if err == io.EOF {
						break
					}
					log.Printf("Error decoding message object in %s: %v. Skipping to next.\n", filePath, err)
					continue
				}

				if msg.Author.ID == "" {
					continue
				}

				displayName := msg.Author.Nickname
				if displayName == "" {
					displayName = msg.Author.Name
				}

				if len(msg.Author.Roles) == 0 {
					entryKey := fmt.Sprintf("%s-NO_ROLE_ASSIGNED", msg.Author.ID)
					mu.Lock()
					if _, exists := uniqueEntries[entryKey]; !exists {
						uniqueEntries[entryKey] = UserRoleEntry{
							UserID:      msg.Author.ID,
							Username:    msg.Author.Name,
							DisplayName: displayName,
							RoleID:      "",
							RoleName:    "",
						}
					}
					mu.Unlock()
				} else {
					for _, role := range msg.Author.Roles {
						roleID := role.ID
						roleName := role.Name

						roleKeyPart := role.ID
						if roleKeyPart == "" {
							roleKeyPart = "EMPTY_ROLE_ID"
						}
						entryKey := fmt.Sprintf("%s-%s", msg.Author.ID, roleKeyPart)

						mu.Lock()
						if _, exists := uniqueEntries[entryKey]; !exists {
							uniqueEntries[entryKey] = UserRoleEntry{
								UserID:      msg.Author.ID,
								Username:    msg.Author.Name,
								DisplayName: displayName,
								RoleID:      roleID,
								RoleName:    roleName,
							}
						}
						mu.Unlock()
					}
				}
			}
			_, err = decoder.Token()
			if err != nil && err != io.EOF {
				log.Printf("Error reading closing bracket for messages array in %s: %v\n", filePath, err)
			}
			break
		} else {
			var dummy interface{}
			if err := decoder.Decode(&dummy); err != nil {
				if err == io.EOF {
					break
				}
				log.Printf("Error skipping value for key '%s' in %s: %v\n", key, filePath, err)
				return
			}
		}
	}
	if !foundMessages {
		log.Printf("Warning: 'messages' key not found in %s\n", filePath)
	}
	log.Printf("Finished processing file: %s\n", filePath)
}

func main() {
	outputFile := flag.String("o", "output.csv", "Output CSV file name")
	flag.Parse()

	jsonFilePaths := flag.Args()
	if len(jsonFilePaths) == 0 {
		log.Fatal("No input JSON files provided. Usage: go run main.go -o output.csv file1.json file2.json ...")
	}

	uniqueEntries := make(map[string]UserRoleEntry)
	var mu sync.Mutex
	var wg sync.WaitGroup

	var allFiles []string
	for _, pathArg := range jsonFilePaths {
		matches, err := filepath.Glob(pathArg)
		if err != nil {
			log.Printf("Error with glob pattern %s: %v\n", pathArg, err)
			continue
		}
		if len(matches) == 0 {
			if _, err := os.Stat(pathArg); err == nil {
				allFiles = append(allFiles, pathArg)
			} else {
				log.Printf("Warning: File or pattern not found: %s\n", pathArg)
			}
		} else {
			allFiles = append(allFiles, matches...)
		}
	}

	if len(allFiles) == 0 {
		log.Fatal("No valid input JSON files found after processing arguments.")
	}

	for _, filePath := range allFiles {
		info, err := os.Stat(filePath)
		if err != nil {
			log.Printf("Error stating file %s: %v. Skipping.\n", filePath, err)
			continue
		}
		if info.IsDir() {
			log.Printf("Skipping directory: %s\n", filePath)
			continue
		}
		if !strings.HasSuffix(strings.ToLower(filePath), ".json") {
			log.Printf("Skipping non-JSON file: %s\n", filePath)
			continue
		}

		wg.Add(1)
		go processJSONFile(filePath, uniqueEntries, &mu, &wg)
	}

	wg.Wait()

	log.Printf("All files processed. Writing to CSV: %s\n", *outputFile)

	csvFile, err := os.Create(*outputFile)
	if err != nil {
		log.Fatalf("Error creating CSV file %s: %v\n", *outputFile, err)
	}
	defer csvFile.Close()

	csvWriter := csv.NewWriter(bufio.NewWriter(csvFile))
	defer csvWriter.Flush()

	headers := []string{"UserID", "Username", "DisplayName", "RoleID", "RoleName"}
	if err := csvWriter.Write(headers); err != nil {
		log.Fatalf("Error writing CSV header: %v\n", err)
	}

	count := 0
	for _, entry := range uniqueEntries {
		record := []string{
			entry.UserID,
			entry.Username,
			entry.DisplayName,
			entry.RoleID,
			entry.RoleName,
		}
		if err := csvWriter.Write(record); err != nil {
			log.Printf("Error writing record to CSV: %v\n", err)
		}
		count++
	}

	log.Printf("Successfully wrote %d unique user-role entries to %s\n", count, *outputFile)
}
