package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"

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
	r.HandleFunc("/admin/recruits", server.RecruitAllGet).Methods("GET")
	r.HandleFunc("/admin/recruits/active", server.RecruitActive).Methods("PUT")
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

	fmt.Println("サーバー起動 : 60012 port で受信")
	// log.Fatal は、異常を検知すると処理の実行を止めてくれる
	log.Fatal(http.ListenAndServe(":60012", c))
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

// Recruitのmembersの構造体
type RecruitsMembers struct {
	Uid      *string `json:"uid,omitempty" dynamodbav:"uid,omitempty"`
	Position *string `json:"position,omitempty" dynamodbav:"position,omitempty"`
}

// sortのインターフェース
func (r AllGetType) Len() int {
	return len(r)
}
func (r AllGetType) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}
func (r AllGetType) Less(i, j int) bool {
	return *r[i].Id > *r[j].Id
}

// ==================== AllGet ===================
type RecruitAllGetResponse struct {
	Id          *int               `json:"id,omitempty" dynamodbav:"id,omitempty"`
	MasterId    *string            `json:"masterId,omitempty" dynamodbav:"masterId,omitempty"`
	Title       *string            `json:"title,omitempty" dynamodbav:"title,omitempty"`
	EventDay    *string            `json:"eventDay,omitempty" dynamodbav:"eventDay,omitempty"`
	Day         *string            `json:"day,omitempty" dynamodbav:"day,omitempty"`
	Organizer   *string            `json:"organizer,omitempty" dynamodbav:"organizer,omitempty"`
	Commit      *string            `json:"commit,omitempty" dynamodbav:"commit,omitempty"`
	Beginner    *string            `json:"beginner,omitempty" dynamodbav:"beginner,omitempty"`
	Message     *string            `json:"message,omitempty" dynamodbav:"message,omitempty"`
	SlackUrl    *string            `json:"slackUrl,omitempty" dynamodbav:"slackUrl,omitempty"`
	TotalMember *string            `json:"totalMember,omitempty" dynamodbav:"totalMember,omitempty"`
	Position    *string            `json:"position,omitempty" dynamodbav:"position,omitempty"`
	Members     *[]RecruitsMembers `json:"members,omitempty" dynamodbav:"members,omitempty"`
	Created     *string            `json:"created,omitempty" dynamodbav:"created,omitempty"`
	Updated     *string            `json:"updated,omitempty" dynamodbav:"updated,omitempty"`
	IsActive    bool               `json:"isActive" dynamodbav:"isActive"`
}
type AllGetType []RecruitAllGetResponse

func (s *Server) RecruitAllGet(w http.ResponseWriter, r *http.Request) {
	tableName := "Recruits"
	param := &dynamodb.ScanInput{
		TableName: aws.String(tableName),
	}
	result, _ := s.db.Scan(param)

	var resRecruit = make(AllGetType, 0)
	dynamodbattribute.UnmarshalListOfMaps(result.Items, &resRecruit)
	sort.Sort(resRecruit)
	j, _ := json.Marshal(resRecruit)
	w.Write(j)

	// 取得値のログ
	fmt.Println(string(j))
}

// ==================== isActive ====================
type RecruitUpdateRequest struct {
	Id       *int `json:"id,omitempty" dynamodbav:"id,omitempty"`
	IsActive bool `json:"isActive" dynamodbav:"isActive"`
}

func (s *Server) RecruitActive(w http.ResponseWriter, r *http.Request) {
	var reqRecruit RecruitAllGetResponse
	json.Unmarshal(StreamToByte(r.Body), &reqRecruit)

	tableName := "Recruits"
	param := &dynamodb.UpdateItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				N: aws.String(strconv.Itoa(*reqRecruit.Id)),
			},
		},
		UpdateExpression: aws.String("set #A = :a"),
		ExpressionAttributeNames: map[string]*string{
			"#A": aws.String("isActive"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":a": {
				BOOL: &reqRecruit.IsActive,
			},
		},
		ReturnValues: aws.String("UPDATED_NEW"),
	}
	_, _ = s.db.UpdateItem(param)

	j, _ := json.Marshal(reqRecruit)
	// 変更値のログ
	fmt.Println(string(j))
}
