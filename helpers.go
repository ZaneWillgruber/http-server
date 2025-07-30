package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

var badWords = [...]string{"kerfuffle", "sharbert", "fornax"}

func respondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	type errorVals struct {
		Error string `json:"error"`
	}

	respBody := errorVals{
		Error: message,
	}

	data, err := json.Marshal(respBody)
	if err != nil {
		fmt.Printf("Error marshalling JSON: %s\n", err)
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(code)
	w.Write(data)
}

func respondWithJSON(w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	data, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("Error marshalling JSON: %s\n", err)
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(code)
	w.Write(data)
}

func recieveJSON(w http.ResponseWriter, r *http.Request, payload any) error {
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&payload)
	if err != nil {
		fmt.Printf("Error decoding parameters: %s\n", err)
		respondWithError(w, 500, "something went wrong")
		return err
	}

	return nil
}

func censorProfaneWords(body string) string {
	split := strings.Split(body, " ")

	for i, word := range split {
		for _, badWord := range badWords {
			if strings.EqualFold(word, badWord) {
				split[i] = "****"
			}
		}
	}

	return strings.Join(split, " ")
}

func mapSlice[S, D any](source []S, converter func(S) D) []D {
	destination := make([]D, 0, len(source))

	for _, item := range source {
		destination = append(destination, converter(item))
	}

	return destination
}
