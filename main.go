package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const apiVersion = "v0.0.1"

type Config struct {
	Schema string `json:"schema"`
	Host   string `json:"host"`
	Port   string `json:"port"`
}

type Configs struct {
	Test    Config `json:"test"`
	Product Config `json:"product"`
}

type QuestionAnswerRequest struct {
	QuestionID       string `json:"questionID" bson:"question_id"`
	UserID           string `json:"userID" bson:"user_id"`
	QuestionNumber   int    `json:"questionNumber" bson:"question_number"`
	Version          string `json:"version" bson:"version"`
	QuestionAnswerID int    `json:"questionAnswerID" bson:"question_answer_id"`
	QuestionValue    string `json:"questionValue" bson:"question_value"`
	Lang             string `json:"lang" bson:"lang"`
}

type QuestionAnswersRequest []QuestionAnswerRequest

type User struct {
	ID              string                 `bson:"id"`
	QuestionAnswers QuestionAnswersRequest `bson:"question_answers"`
}

type Users []User

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

type QuestionRequest struct {
	Version string `json:"version" bson:"version"`
	UserID  string `json:"userID" bson:"user_id"`
	Value   string `json:"value" bson:"value"`
	Lang    string `json:"lang" bson:"lang"`
}

type QuestionsRequest []QuestionRequest

type QuestionResponse struct {
	Version        string `json:"version"`
	QuestionID     string `json:"questionID"`
	QuestionNumber int    `json:"questionNumber"`
	QuestionJA     string `json:"questionJA"`
	QuestionEN     string `json:"questionEN"`
	URL            string `json:"url"`
	Title          string `json:"title"`
	Description    string `json:"description"`
}

type QuestionsResponse []QuestionResponse

type ErrorResponse struct {
	Message string `json:"message"`
}

type ErrorsResponse []ErrorResponse

type Favorite struct {
	ID          string `bson:"_id"`
	Version     string `json:"version"`
	Href        string `bson:"href"`
	IsChecked   bool   `bson:"is_checked"`
	Description string `bson:"description"`
	Title       string `bson:"title"`
}

type Faovrites []Favorite

type Word struct {
	ID          string `bson:"_id"`
	SectionName string `bson:"section_name"`
	Type        string `bson:"type"`
	Lang        string `bson:"lang"`
	Href        string `bson:"href"`
	Count       int    `bson:"count"`
	Value       string `bson:"value"`
	UpperValue  string `bson:"upper_value"`
	JPNickname  string `bson:"jp_nickname"`
}

type Words []Word

type Question struct {
	ID                   string `bson:"_id"`
	QuestionJA           string `bson:"question_ja"`
	QuestionEN           string `bson:"question_en"`
	QuestionSeedEN       string `bson:"question_seed_en"`
	QuestionSeedJA       string `bson:"question_seed_ja"`
	QuestionSeedType     string `bson:"question_seed_type"`
	TranslatedFromJAToEN string `bson:"translated_from_ja_to_en"`
}

type Questions []Question

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

func CORSforOptions(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
	(*w).Header().Set("Access-Control-Allow-Headers", "Content-Type")
	(*w).WriteHeader(204)
}

func removeDuplicate(args []string) []string {
	results := make([]string, 0, len(args))
	encountered := map[string]bool{}
	for i := 0; i < len(args); i++ {
		if !encountered[args[i]] {
			encountered[args[i]] = true
			results = append(results, args[i])
		}
	}
	return results
}

func userQueryRequest() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "OPTIONS":
			CORSforOptions(&w)
			return
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

			// url, err := url.Parse(userQueryRequest.Href)
			// if err != nil {
			// 	fmt.Println(err)
			// }

			// queries := url.Query()

			// keys, ok := queries["q"]

			// if !ok || len(keys[0]) < 1 {
			// 	fmt.Println(keys, " is missing")
			// 	return
			// }

			// searchValue := keys[0]

			// userQueryRequest.Value = searchValue

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
		case "OPTIONS":
			CORSforOptions(&w)
			return
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

func questionOrder(questionRequest QuestionRequest, orderQuestionID int) (error, QuestionsResponse) {

	searchValue := questionRequest.Value

	splitedSearchValue := regexp.MustCompile("[ 　	]").Split(searchValue, -1)

	fmt.Println(splitedSearchValue, len(searchValue))

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		return err, nil
	}
	defer client.Disconnect(ctx)

	words_collection := client.Database("oborobot").Collection("word")

	filter := bson.M{}
	cur, err := words_collection.Find(ctx, filter)
	if err != nil {
		return err, nil
	}

	words := Words{}

	err = cur.All(ctx, &words)
	if err != nil {
		return err, nil
	}

	urlList := []string{}
	for i := 0; i < len(splitedSearchValue); i++ {
		splitedSearchValueEl := splitedSearchValue[i]
		for j := 0; j < len(words); j++ {
			word := words[j]
			if word.UpperValue == strings.ToUpper(splitedSearchValueEl) {
				haveValue := false
				for k := 0; k < len(urlList); k++ {
					if urlList[k] == word.Href {
						haveValue = true
					}
				}
				if haveValue == false && !strings.Contains(word.Href, "www.google.com/search") {
					urlList = append(urlList, word.Href)
				}
			}
		}
	}

	if len(urlList) == 0 {
		return errors.New("ごめんなさい｡対応できないかも..."), nil
	}

	urlFilterWords := Words{}
	keyWords := []string{}

	for i := 0; i < len(urlList); i++ {
		findOptions := options.Find()
		findOptions.SetSort(bson.D{{"count", -1}})
		urlFilter := bson.M{"href": urlList[i]}
		cur, err = words_collection.Find(ctx, urlFilter, findOptions)
		if err != nil {
			return err, nil
		}
		temp_urlFilterWords := Words{}
		err = cur.All(ctx, &temp_urlFilterWords)
		if err != nil {
			return err, nil
		}
		// fmt.Println(temp_urlFilterWords)
		for j := 0; j < len(temp_urlFilterWords); j++ {
			temp_url_filter_word := temp_urlFilterWords[j]
			if temp_url_filter_word.Type != "Verb" {
				urlFilterWords = append(urlFilterWords, temp_url_filter_word)
				haveValue := false
				for k := 0; k < len(splitedSearchValue); k++ {
					if splitedSearchValue[k] == temp_url_filter_word.Value {
						haveValue = true
					}
				}
				if haveValue == false {
					keyWords = append(keyWords, temp_url_filter_word.Value)
				}
			}
		}
	}

	// fmt.Println(urlFilterWords, len(urlFilterWords))
	// fmt.Println(keyWords, 111)

	// Goの場合 urlList
	// [https://frasco.io/why-we-switched-from-python-to-go-19581e27de7c https://teratail.com/questions/93783 http://www.spinute.org/go-by-example/url-parsing.html]

	maxLoopNumber := 10
	questionCount := 5

	// select_keywords_success := false
	selectKeywords := []string{}
	for i := 0; i < maxLoopNumber; i++ {

		selectKeywords = []string{}
		for j := 0; j < questionCount; j++ {
			rand.Seed(time.Now().UnixNano())
			randomValue := rand.Intn(len(keyWords))
			selectKeywords = append(selectKeywords, keyWords[randomValue])
		}

		if len(removeDuplicate(selectKeywords)) == questionCount {
			// select_keywords_success = true
			break
		}
	}

	// fmt.Println(selectKeywords)

	questions_collection := client.Database("oborobot").Collection("question")
	filter = bson.M{}
	cur, err = questions_collection.Find(ctx, filter)
	if err != nil {
		return err, nil
	}

	questions := Questions{}

	err = cur.All(ctx, &questions)
	if err != nil {
		return err, nil
	}

	// fmt.Println(questions)

	suggest_question := []string{}
	group_suggest_questions := []map[string]string{}
	for i := 0; i < len(questions); i++ {
		question := questions[i]
		haveSelectKeyword := false
		for j := 0; j < len(selectKeywords); j++ {
			selectKeyword := selectKeywords[j]
			// if strings.Contains(question.Question, selectKeyword) || strings.Contains(question.QuestionSeedEN, selectKeyword) || strings.Contains(question.QuestionSeedJA, selectKeyword) && question.QuestionSeedType != "Verb" && question.Lang == questionRequest.Lang {
			// 	haveSelectKeyword = true
			// 	break
			// }

			// TODO: あっているか要確認
			if (strings.ToUpper(question.QuestionSeedEN) == strings.ToUpper(selectKeyword) || strings.ToUpper(question.QuestionSeedJA) == strings.ToUpper(selectKeyword)) && question.QuestionSeedType != "Verb" {

				haveString := false
				for k := 0; k < len(suggest_question); k++ {
					suggest_question_value := suggest_question[k]
					if strings.Contains(strings.ToUpper(suggest_question_value), strings.ToUpper(selectKeyword)) {
						haveString = true
					}
				}
				if haveString == false {
					haveSelectKeyword = true
					break
				}

			}
		}

		if haveSelectKeyword == true {
			suggest_question = append(suggest_question, question.QuestionJA)
			suggest_question = append(suggest_question, question.QuestionEN)

			suggest_questions := map[string]string{"question_id": question.ID, "question_ja": question.QuestionJA, "question_en": question.QuestionEN}
			group_suggest_questions = append(group_suggest_questions, suggest_questions)
			// fmt.Println(question.QuestionJA, question.QuestionEN)
		}
	}
	fmt.Println(group_suggest_questions)
	fmt.Println(urlList)
	// fmt.Println(suggest_question)

	rand.Seed(time.Now().UnixNano())
	randomValue := rand.Intn(len(urlList))
	suggest_url := urlList[randomValue]

	randomValue = rand.Intn(len(group_suggest_questions))
	suggest_ja_en_question := group_suggest_questions[randomValue]

	favorites_collection := client.Database("oborobot").Collection("favorite")

	filter = bson.M{"href": suggest_url}
	result := favorites_collection.FindOne(ctx, filter)

	favorite := Favorite{}
	err = result.Decode(&favorite)
	suggest_url_title := ""
	suggest_url_description := ""

	// if err != nil {
	// 	responseErrorJSON(w, http.StatusInternalServerError, err.Error())
	// 	return
	// }
	if err == nil {
		suggest_url_title = favorite.Title
		suggest_url_description = favorite.Description
	}

	// fmt.Println(suggest_question, suggest_url, suggest_url_title)

	questionsResponse := QuestionsResponse{
		QuestionResponse{
			Version:        apiVersion,
			QuestionID:     suggest_ja_en_question["question_id"], // TODO: これ推測できる可能性がある｡
			QuestionNumber: orderQuestionID,
			QuestionJA:     suggest_ja_en_question["question_ja"],
			QuestionEN:     suggest_ja_en_question["question_en"],
			URL:            suggest_url,
			Title:          suggest_url_title,
			Description:    suggest_url_description,
		},
	}

	return nil, questionsResponse
}

func questionRequest() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "OPTIONS":
			CORSforOptions(&w)
			return
		case "POST":
			requestBody, err := ioutil.ReadAll(r.Body)
			defer r.Body.Close()
			if err != nil {
				responseErrorJSON(w, http.StatusInternalServerError, err.Error())
				return
			}

			questionRequest := QuestionRequest{
				Version: apiVersion,
			}

			err = json.Unmarshal(requestBody, &questionRequest)
			if err != nil {
				responseErrorJSON(w, http.StatusInternalServerError, err.Error())
				return
			}

			fmt.Println(questionRequest)

			if len(questionRequest.UserID) != 36 {
				responseErrorJSON(w, http.StatusInternalServerError, "requestBody don't have 'userID'.")
				return
			} else if len(questionRequest.Value) == 0 {
				responseErrorJSON(w, http.StatusInternalServerError, "'questionValue is invalid.'")
				return
			} else if questionRequest.Lang != "ja" && questionRequest.Lang != "en" {
				responseErrorJSON(w, http.StatusInternalServerError, "requestBody don't have 'lang'.")
				return
			}

			err, questionsResponse := questionOrder(questionRequest, 1)
			if err != nil {
				responseErrorJSON(w, http.StatusInternalServerError, err.Error())
			}

			// users := Users{}
			ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
			client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
			if err != nil {
				responseErrorJSON(w, http.StatusInternalServerError, err.Error())
			}
			defer client.Disconnect(ctx)

			user := User{}
			userCollection := client.Database("oborobot").Collection("user")
			idFilter := bson.M{"id": questionRequest.UserID}
			userResult := userCollection.FindOne(ctx, idFilter)
			err = userResult.Decode(&user)

			if err != nil { // userが存在しない
				user = User{
					ID:              questionRequest.UserID,
					QuestionAnswers: QuestionAnswersRequest{
						// QuestionAnswerRequest{},
					},
				}
				_, err = userCollection.InsertOne(ctx, user)
				if err != nil {
					responseErrorJSON(w, http.StatusInternalServerError, err.Error())
				}
			}

			responseJSON(w, http.StatusOK, questionsResponse)
			return

		default:
			responseErrorJSON(w, http.StatusMethodNotAllowed, "Sorry, only POST method is supported.")
			return
		}

		// unreach
	}

}

func questionAnswerRequest() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "OPTIONS":
			CORSforOptions(&w)
			return
		case "POST":
			requestBody, err := ioutil.ReadAll(r.Body)
			defer r.Body.Close()
			if err != nil {
				responseErrorJSON(w, http.StatusInternalServerError, err.Error())
				return
			}
			questionAnswer := QuestionAnswerRequest{
				Version: apiVersion,
			}

			err = json.Unmarshal(requestBody, &questionAnswer)
			if err != nil {
				responseErrorJSON(w, http.StatusInternalServerError, err.Error())
				return
			}

			if len(questionAnswer.UserID) != 36 {
				responseErrorJSON(w, http.StatusInternalServerError, "requestBody don't have 'userID'.")
				return
			} else if len(questionAnswer.QuestionID) != 24 {
				responseErrorJSON(w, http.StatusInternalServerError, "'questionID is invalid.'")
				return
			} else if len(questionAnswer.QuestionValue) == 0 {
				responseErrorJSON(w, http.StatusInternalServerError, "'questionValue is invalid.'")
				return
			} else if questionAnswer.QuestionNumber <= 0 {
				responseErrorJSON(w, http.StatusInternalServerError, "'questionNumber is invalid.'")
				return
			} else if questionAnswer.QuestionAnswerID < 1 || questionAnswer.QuestionAnswerID > 5 {
				responseErrorJSON(w, http.StatusInternalServerError, "'questionAnswerID is invalid.'")
				return
			} else if questionAnswer.Lang != "ja" && questionAnswer.Lang != "en" {
				responseErrorJSON(w, http.StatusInternalServerError, "requestBody don't have 'lang'.")
				return
			}

			// url, err := url.Parse(userQueryRequest.Href)
			// if err != nil {
			// 	fmt.Println(err)
			// }

			// queries := url.Query()

			// keys, ok := queries["q"]

			// if !ok || len(keys[0]) < 1 {
			// 	fmt.Println(keys, " is missing")
			// 	return
			// }

			// searchValue := keys[0]

			// userQueryRequest.Value = searchValue

			// fmt.Println(searchValue)

			ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
			client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
			if err != nil {
				responseErrorJSON(w, http.StatusInternalServerError, err.Error())
				return
			}
			defer client.Disconnect(ctx)

			user_collection := client.Database("oborobot").Collection("user")
			filter := bson.M{"id": questionAnswer.UserID}
			fmt.Println(questionAnswer)
			change := bson.M{
				"$push": bson.M{
					"question_answers": bson.M{
						"question_id":        questionAnswer.QuestionID,
						"question_number":    questionAnswer.QuestionNumber,
						"question_answer_id": questionAnswer.QuestionAnswerID,
						"lang":               questionAnswer.Lang,
					},
				},
			}
			_, err = user_collection.UpdateOne(ctx, filter, change)
			if err != nil {
				responseErrorJSON(w, http.StatusInternalServerError, err.Error())
			}

			questionRequest := QuestionRequest{
				UserID: questionAnswer.UserID,
				Value:  questionAnswer.QuestionValue,
				Lang:   questionAnswer.Lang,
			}

			err, questionsResponse := questionOrder(questionRequest, questionAnswer.QuestionNumber+1)
			if err != nil {
				responseErrorJSON(w, http.StatusInternalServerError, err.Error())
			}

			responseJSON(w, http.StatusOK, questionsResponse)
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

	configsJSON, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	configs := Configs{}

	err = json.Unmarshal(configsJSON, &configs)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	envType := flag.Arg(0)

	const envTypeTest = "test"
	const envTypeProduct = "product"

	config := Config{}
	if envType == envTypeTest {
		config = configs.Test
	} else if envType == envTypeProduct {
		config = configs.Product
	} else {
		log.Println("config-type is invalid.")
		os.Exit(1)
	}

	apiEndpointName := "/api"
	userEndpointName := apiEndpointName + "/user"
	http.HandleFunc(userEndpointName+"/query", userQueryRequest())
	http.HandleFunc(userEndpointName+"/favorite", userFavoriteRequest())

	questionEndpointName := apiEndpointName + "/question"
	http.HandleFunc(questionEndpointName, questionRequest())

	questionAnswerEndpointName := apiEndpointName + "/questionAnswer"
	http.HandleFunc(questionAnswerEndpointName, questionAnswerRequest())

	schema := config.Schema
	host := config.Host
	port := config.Port
	addr := host + ":" + port
	fmt.Println("LISTEN: ", schema+"://"+addr)

	// 証明書の作成参考: https://ozuma.hatenablog.jp/entry/20130511/1368284304
	if envType == envTypeTest {
		err = http.ListenAndServeTLS(addr, "cert_key/test/cert.pem", "cert_key/test/key.pem", nil)
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
	} else if envType == envTypeProduct {
		err = http.ListenAndServeTLS(addr, "cert_key/product/cert.pem", "cert_key/product/key.pem", nil)
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
	} else {
		log.Println("config-type is invalid.")
		os.Exit(1)
	}

}
