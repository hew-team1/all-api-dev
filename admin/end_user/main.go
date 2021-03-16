package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {
	sess, _ := session.NewSession(&aws.Config{
		Region:      aws.String(os.Getenv("REGION")),
		Endpoint:    aws.String(os.Getenv("ENDPOINT_DB")),
		Credentials: credentials.NewStaticCredentials(os.Getenv("AWS_ACCESS_KEY_ID"), os.Getenv("AWS_SECRET_ACCESS_KEY"), ""),
	})
	db := dynamodb.New(sess)
	server := NewServer(db)

	r := mux.NewRouter()
	r.HandleFunc("/admin/users", server.UserAllGet).Methods("GET")
	r.HandleFunc("/admin/users/active", server.UserActive).Methods("PUT")
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedHeaders: []string{"*"},
		AllowedMethods: []string{
			http.MethodHead,
			http.MethodGet,
			http.MethodPost,
			http.MethodPatch,
			http.MethodPut,
			http.MethodDelete,
		},
	}).Handler(r)

	fmt.Println("サーバー起動 : 60011 port で受信")

	// log.Fatal は、異常を検知すると処理の実行を止めてくれる
	log.Fatal(http.ListenAndServe(":60011", c))
}

// io.Readerをbyteのスライスに変換
func StreamToByte(stream io.Reader) []byte {
	buf := new(bytes.Buffer)
	buf.ReadFrom(stream)
	return buf.Bytes()
}

func NewServer(db *dynamodb.DynamoDB) *Server {
	return &Server{
		db: db,
	}
}

type Server struct {
	db *dynamodb.DynamoDB
}

// ==================== ALLGet ====================

type UserAllGetResponse struct {
	Uid      *string `json:"uid,omitempty" dynamodbav:"uid,omitempty"`
	Name     *string `json:"name,omitempty" dynamodbav:"name,omitempty"`
	Email    *string `json:"email,omitempty" dynamodbav:"email,omitempty"`
	Created  *string `json:"created,omitempty" dynamodbav:"created,omitempty"`
	Updated  *string `json:"updated,omitempty" dynamodbav:"updated,omitempty"`
	IsLogin  bool    `json:"isLogin" dynamodbav:"isLogin"`
	IsActive bool    `json:"isActive" dynamodbav:"isActive"`
}

func (s *Server) UserAllGet(w http.ResponseWriter, r *http.Request) {

	tableName := "EndUsers"

	param := &dynamodb.ScanInput{
		TableName: aws.String(tableName),
	}
	result, _ := s.db.Scan(param)

	var resUser []UserAllGetResponse
	dynamodbattribute.UnmarshalListOfMaps(result.Items, &resUser)
	j, _ := json.Marshal(resUser)
	w.Write(j)

	// 取得値のログ
	fmt.Println(string(j))
}

// ==================== isActive ====================
type UserUpdateRequest struct {
	Uid      *string `json:"uid,omitempty" dynamodbav:"uid,omitempty"`
	IsActive bool    `json:"isActive" dynamodbav:"isActive"`
}

func (s *Server) UserActive(w http.ResponseWriter, r *http.Request) {
	var reqUser UserUpdateRequest
	json.Unmarshal(StreamToByte(r.Body), &reqUser)

	tableName := "EndUsers"
	param := &dynamodb.UpdateItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"uid": {
				S: aws.String(*reqUser.Uid),
			},
		},
		UpdateExpression: aws.String("set #A = :a"),
		ExpressionAttributeNames: map[string]*string{
			"#A": aws.String("isActive"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":a": {
				BOOL: &reqUser.IsActive,
			},
		},
	}
	_, _ = s.db.UpdateItem(param)

	j, _ := json.Marshal(reqUser)
	// 変更値のログ
	fmt.Println(string(j))
}
