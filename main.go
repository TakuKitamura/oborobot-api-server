package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

type UserFavoriteRequest struct {
	Version string `json:"version"`
	Href    string `json:"href"`
}

type UserFavoritesRequest []UserFavoriteRequest

type UserFavoriteResponse struct {
	Version string `json:"version"`
}

type UserFavoritesResponse []UserFavoriteResponse

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
	fmt.Println(message)
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

func userQueryRequest() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			requestBody, err := ioutil.ReadAll(r.Body)
			defer r.Body.Close()
			if err != nil {
				responseErrorJSON(w, http.StatusInternalServerError, err.Error())
				return
			}
			userQueryRequest := UserQueryRequest{
				Version: apiVersion,
			}

			err = json.Unmarshal(requestBody, &userQueryRequest)
			if err != nil {
				responseErrorJSON(w, http.StatusInternalServerError, err.Error())
				return
			}

			fmt.Println(userQueryRequest)

			ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
			client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
			if err != nil {
				responseErrorJSON(w, http.StatusInternalServerError, err.Error())
				return
			}
			defer client.Disconnect(ctx)

			collection := client.Database("oborobot").Collection("query")
			_, err = collection.InsertOne(ctx, userQueryRequest)
			if err != nil {
				responseErrorJSON(w, http.StatusInternalServerError, err.Error())
			}

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

}

func userFavoriteRequest() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			requestBody, err := ioutil.ReadAll(r.Body)
			defer r.Body.Close()
			if err != nil {
				responseErrorJSON(w, http.StatusInternalServerError, err.Error())
				return
			}

			userFavoriteRequest := UserFavoriteRequest{
				Version: apiVersion,
			}

			err = json.Unmarshal(requestBody, &userFavoriteRequest)
			if err != nil {
				responseErrorJSON(w, http.StatusInternalServerError, err.Error())
				return
			}

			fmt.Println(userFavoriteRequest)

			ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
			client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
			if err != nil {
				responseErrorJSON(w, http.StatusInternalServerError, err.Error())
				return
			}
			defer client.Disconnect(ctx)

			collection := client.Database("oborobot").Collection("favorite")
			_, err = collection.InsertOne(ctx, userFavoriteRequest)
			if err != nil {
				responseErrorJSON(w, http.StatusInternalServerError, err.Error())
			}

			userFavoritesResponse := UserFavoritesResponse{
				UserFavoriteResponse{
					Version: apiVersion,
				},
			}

			responseJSON(w, http.StatusOK, userFavoritesResponse)
			return

		default:
			responseErrorJSON(w, http.StatusMethodNotAllowed, "Sorry, only POST method is supported.")
			return
		}

		// unreach
	}

}

func main() {
	apiEndpointName := "/api"
	userEndpointName := apiEndpointName + "/user"
	http.HandleFunc(userEndpointName+"/query", userQueryRequest())
	http.HandleFunc(userEndpointName+"/favorite", userFavoriteRequest())
	http.ListenAndServe(":3000", nil)
}
