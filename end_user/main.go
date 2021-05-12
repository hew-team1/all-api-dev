package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

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
	r.HandleFunc("/users", server.UserAllGet).Methods("GET")
	r.HandleFunc("/users", server.UserCreate).Methods("POST")
	r.HandleFunc("/users/in-posts", server.InPostsGet).Methods("GET")
	r.HandleFunc("/users/in-join", server.InJoin).Methods("GET")
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

	fmt.Println("サーバー起動 :80 port で受信")

	// log.Fatal は、異常を検知すると処理の実行を止めてくれる
	log.Fatal(http.ListenAndServe(":60001", c))
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
	Uid   *string `json:"uid,omitempty" dynamodbav:"uid,omitempty"`
	Name  *string `json:"name,omitempty" dynamodbav:"name,omitempty"`
	Email *string `json:"email,omitempty" dynamodbav:"email,omitempty"`
}

func (s *Server) UserAllGet(w http.ResponseWriter, r *http.Request) {
	tableName := "EndUsers"
	active := true
	param := &dynamodb.ScanInput{
		TableName:        aws.String(tableName),
		FilterExpression: aws.String("#A = :a"),
		ExpressionAttributeNames: map[string]*string{
			"#A": aws.String("isActive"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":a": {
				BOOL: &active,
			},
		},
	}
	result, _ := s.db.Scan(param)

	var resUser = make([]UserAllGetResponse, 0)
	dynamodbattribute.UnmarshalListOfMaps(result.Items, &resUser)
	j, _ := json.Marshal(resUser)
	w.Write(j)

	// 取得値のログ
	fmt.Println(string(j))
}

// ==================== Create ====================
type UserCreateRequest struct {
	Uid      *string `json:"uid,omitempty" dynamodbav:"uid,omitempty"`
	Name     *string `json:"name,omitempty" dynamodbav:"name,omitempty"`
	Email    *string `json:"email,omitempty" dynamodbav:"email,omitempty"`
	Created  *string `json:"created,omitempty" dynamodbav:"created,omitempty"`
	Updated  *string `json:"updated,omitempty" dynamodbav:"updated,omitempty"`
	IsLogin  bool    `json:"isLogin" dynamodbav:"isLogin"`
	IsActive bool    `json:"isActive" dynamodbav:"isActive"`
}

func (s *Server) UserCreate(w http.ResponseWriter, r *http.Request) {
	nowTime := time.Now().UTC().In(
		time.FixedZone("Asia/Tokyo", 9*60*60),
	).Format("2006-01-02 15:04")

	var reqUser UserCreateRequest
	json.Unmarshal(StreamToByte(r.Body), &reqUser)

	reqUser.Created = &nowTime
	reqUser.Updated = &nowTime
	reqUser.IsActive = true
	reqUser.IsLogin = true

	av, err := dynamodbattribute.MarshalMap(reqUser)
	tableName := "EndUsers"
	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(tableName),
	}

	_, err = s.db.PutItem(input)
	if err != nil {
		fmt.Println("Got error calling PutItem:")
		fmt.Println(err.Error())
		os.Exit(1)
	}
	j, _ := json.Marshal(reqUser)

	// 作成値のログ
	fmt.Println(string(j))
}

// ==================== inPosts ====================
type InPostsGetResponse struct {
	Id          *int    `json:"id,omitempty" dynamodbav:"id,omitempty"`
	MasterId    *string `json:"masterId,omitempty" dynamodbav:"masterId,omitempty"`
	Title       *string `json:"title,omitempty" dynamodbav:"title,omitempty"`
	EventDay    *string `json:"eventDay,omitempty" dynamodbav:"eventDay,omitempty"`
	Day         *string `json:"day,omitempty" dynamodbav:"day,omitempty"`
	Organizer   *string `json:"organizer,omitempty" dynamodbav:"organizer,omitempty"`
	Commit      *string `json:"commit,omitempty" dynamodbav:"commit,omitempty"`
	Beginner    *string `json:"beginner,omitempty" dynamodbav:"beginner,omitempty"`
	Message     *string `json:"message,omitempty" dynamodbav:"message,omitempty"`
	SlackUrl    *string `json:"slackUrl,omitempty" dynamodbav:"slackUrl,omitempty"`
	TotalMember *string `json:"totalMember,omitempty" dynamodbav:"totalMember,omitempty"`
	Position    *string `json:"position,omitempty" dynamodbav:"position,omitempty"`
	Members     *[]struct {
		Uid      *string `json:"uid,omitempty" dynamodbav:"uid,omitempty"`
		Position *string `json:"position,omitempty" dynamodbav:"position,omitempty"`
	} `json:"members,omitempty" dynamodbav:"members,omitempty"`
	Created *string `json:"created,omitempty" dynamodbav:"created,omitempty"`
	Updated *string `json:"updated,omitempty" dynamodbav:"updated,omitempty"`
}

func (s *Server) InPostsGet(w http.ResponseWriter, r *http.Request) {
	uid := r.Header.Get("uid")

	tableName := "Recruits"
	active := true
	param := &dynamodb.ScanInput{
		TableName:        aws.String(tableName),
		FilterExpression: aws.String("#M = :m AND #A = :a"),
		ExpressionAttributeNames: map[string]*string{
			"#M": aws.String("masterId"),
			"#A": aws.String("isActive"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":m": {
				S: aws.String(uid),
			},
			":a": {
				BOOL: &active,
			},
		},
	}
	result, _ := s.db.Scan(param)

	var resInPosts = make([]InPostsGetResponse, 0)
	dynamodbattribute.UnmarshalListOfMaps(result.Items, &resInPosts)
	j, _ := json.Marshal(resInPosts)
	w.Write(j)

	// 返却値のログ
	fmt.Println(string(j))
}

// ====================InJoin ====================
type InJoinGetResponse struct {
	Id          *int    `json:"id,omitempty" dynamodbav:"id,omitempty"`
	MasterId    *string `json:"masterId,omitempty" dynamodbav:"masterId,omitempty"`
	Title       *string `json:"title,omitempty" dynamodbav:"title,omitempty"`
	EventDay    *string `json:"eventDay,omitempty" dynamodbav:"eventDay,omitempty"`
	Day         *string `json:"day,omitempty" dynamodbav:"day,omitempty"`
	Organizer   *string `json:"organizer,omitempty" dynamodbav:"organizer,omitempty"`
	Commit      *string `json:"commit,omitempty" dynamodbav:"commit,omitempty"`
	Beginner    *string `json:"beginner,omitempty" dynamodbav:"beginner,omitempty"`
	Message     *string `json:"message,omitempty" dynamodbav:"message,omitempty"`
	SlackUrl    *string `json:"slackUrl,omitempty" dynamodbav:"slackUrl,omitempty"`
	TotalMember *string `json:"totalMember,omitempty" dynamodbav:"totalMember,omitempty"`
	Position    *string `json:"position,omitempty" dynamodbav:"position,omitempty"`
	Members     *[]struct {
		Uid      *string `json:"uid,omitempty" dynamodbav:"uid,omitempty"`
		Position *string `json:"position,omitempty" dynamodbav:"position,omitempty"`
	} `json:"members,omitempty" dynamodbav:"members,omitempty"`
	Created *string `json:"created,omitempty" dynamodbav:"created,omitempty"`
	Updated *string `json:"updated,omitempty" dynamodbav:"updated,omitempty"`
}

func (s *Server) InJoin(w http.ResponseWriter, r *http.Request) {
	uid := r.Header.Get("uid")

	tableName := "Recruits"
	active := true

	param := &dynamodb.ScanInput{
		TableName:        aws.String(tableName),
		FilterExpression: aws.String("#A = :a"),
		ExpressionAttributeNames: map[string]*string{
			"#A": aws.String("isActive"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":a": {
				BOOL: &active,
			},
		},
	}
	result, _ := s.db.Scan(param)

	var resInJoin = make([]InJoinGetResponse, 0)
	var allRecruit = make([]InJoinGetResponse, 0)
	dynamodbattribute.UnmarshalListOfMaps(result.Items, &allRecruit)

	// membersにuidがある場合はresInJoinに追加
	for _, row := range allRecruit {
		members := *row.Members
		for _, member := range members {
			if uid == *member.Uid {
				resInJoin = append(resInJoin, row)
			}
		}
	}

	j, _ := json.Marshal(resInJoin)
	w.Write(j)

	//　返却値のログ
	fmt.Println(string(j))
}
