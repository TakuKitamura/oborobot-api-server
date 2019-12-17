package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const apiVersion = "v0.0.1"

type UserQueryRequest struct {
	Version     string `json:"version"`
	Href        string `json:"href"`
	SearchValue string `json:"searchValue"`
}

type UserQueriesRequest []UserQueryRequest

type UserQueryResponse struct {
	Version string `json:"version"`
}

type UserQueriesResponse []UserQueryResponse

type ErrorResponse struct {
	Message string `json:"message"`
}

type ErrorsResponse []ErrorResponse

func responseJSON(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}

func responseErrorJSON(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	errorsResponse := ErrorsResponse{
		ErrorResponse{
			Message: message,
		},
	}
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(errorsResponse)
}

func userQueryRequest(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case "POST":
		requestBody, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			responseErrorJSON(w, http.StatusInternalServerError, err.Error())
			return
		}

		userQueriesRequest := UserQueriesRequest{
			UserQueryRequest{
				Version: apiVersion,
			},
		}

		err = json.Unmarshal(requestBody, &userQueriesRequest)
		if err != nil {
			responseErrorJSON(w, http.StatusInternalServerError, err.Error())
			return
		}

		fmt.Println(userQueriesRequest)

		userQueriesResponse := UserQueriesResponse{
			UserQueryResponse{
				Version: apiVersion,
			},
		}

		responseJSON(w, http.StatusOK, userQueriesResponse)
		return

	default:
		responseErrorJSON(w, http.StatusMethodNotAllowed, "Sorry, only POST method is supported.")
		return
	}

	// unreach
}

func main() {
	apiEndpointName := "/api"
	userEndpointName := apiEndpointName + "/user"
	http.HandleFunc(userEndpointName+"/query", userQueryRequest)
	http.ListenAndServe(":3000", nil)
}
