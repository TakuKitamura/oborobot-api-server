package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const apiVersion = "v0.0.1"

type UserQueryRequest struct {
	Version string `json:"version" bson:"version"`
	Href    string `json:"href" bson:"href"`
	// Lang        string `json:"lang" bson:"lang"`
	SearchValue string `json:"searchValue" bson:"search_value"`
	IsChecked   bool   `json:"isChecked" bson:"is_checked"`
}

type UserQueriesRequest []UserQueryRequest

type UserQueryResponse struct {
	Version string `json:"version"`
}

type UserQueriesResponse []UserQueryResponse

type UserFavoriteRequest struct {
	Version string `json:"version" bson:"version"`
	Href    string `json:"href" bson:"href"`
	// Lang      string `json:"lang" bson:"lang"`
	IsChecked bool `json:"isChecked" bson:"is_checked"`
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

			url, err := url.Parse(userQueryRequest.Href)
			if err != nil {
				fmt.Println(err)
			}

			queries := url.Query()

			keys, ok := queries["q"]

			if !ok || len(keys[0]) < 1 {
				fmt.Println(keys, " is missing")
				return
			}

			searchValue := keys[0]

			userQueryRequest.SearchValue = searchValue

			// fmt.Println(searchValue)

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
	flag.Parse()
	args := flag.Args()

	if len(args) != 1 {
		fmt.Println("args are invalid.")
		return
	}

	apiEndpointName := "/api"
	userEndpointName := apiEndpointName + "/user"
	http.HandleFunc(userEndpointName+"/query", userQueryRequest())
	http.HandleFunc(userEndpointName+"/favorite", userFavoriteRequest())

	ip := flag.Arg(0)
	err := http.ListenAndServeTLS(ip+":3000", "cert_key/cert.pem", "cert_key/key.pem", nil)
	fmt.Println(err)
}
