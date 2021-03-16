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
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {
	var cfgs aws.Config

	cfgs.Region = aws.String(os.Getenv("REGION"))
	cfgs.Endpoint = aws.String(os.Getenv("ENDPOINT_SES"))
	cfgs.Credentials = credentials.NewStaticCredentials(os.Getenv("AWS_ACCESS_KEY_ID"), os.Getenv("AWS_SECRET_ACCESS_KEY"), "")

	sess, _ := session.NewSession(&cfgs)
	ses := ses.New(sess)

	cfgs.Endpoint = aws.String(os.Getenv("ENDPOINT_DB"))
	sess, _ = session.NewSession(&cfgs)
	db := dynamodb.New(sess)

	server := NewServer(db, ses)

	r := mux.NewRouter()
	r.HandleFunc("/recruits", server.RecruitAllGet).Methods("GET")
	r.HandleFunc("/recruits", server.RecruitCreate).Methods("POST")
	r.HandleFunc("/recruits/{id}", server.RecruitGet).Methods("GET")
	r.HandleFunc("/recruits/{id}/members", server.MemberAdd).Methods("PUT")
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
	log.Fatal(http.ListenAndServe(":80", c))
}

// io.Readerをbyteのスライスに変換
func StreamToByte(stream io.Reader) []byte {
	buf := new(bytes.Buffer)
	buf.ReadFrom(stream)
	return buf.Bytes()
}

func NewServer(db *dynamodb.DynamoDB, ses *ses.SES) *Server {
	return &Server{
		db:  db,
		ses: ses,
	}
}

type Server struct {
	db  *dynamodb.DynamoDB
	ses *ses.SES
}

// Recruitのmembersの構造体
type RecruitsMembers struct {
	Uid      *string `json:"uid,omitempty" dynamodbav:"uid,omitempty"`
	Position *string `json:"position,omitempty" dynamodbav:"position,omitempty"`
}

// ==================== Count ====================
func (s *Server) Increment(key string) (max int) {
	tableName := "AtomicCounter"
	param := &dynamodb.UpdateItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"countKey": {
				S: aws.String(key),
			},
		},
		UpdateExpression: aws.String("add #col :incr"),
		ExpressionAttributeNames: map[string]*string{
			"#col": aws.String("countNumber"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":incr": {
				N: aws.String("1"),
			},
		},
		ReturnValues: aws.String("UPDATED_NEW"),
	}
	newCnt, err := s.db.UpdateItem(param)
	if err != nil {
		fmt.Println("Got error calling PutItem:")
		fmt.Println(err.Error())
		os.Exit(1)
	}
	max, _ = strconv.Atoi(*newCnt.Attributes["countNumber"].N)
	return max
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

// ==================== AllGet ====================
type RecruitAllGetResponse struct {
	Id          *int              `json:"id,omitempty" dynamodbav:"id,omitempty"`
	MasterId    *string           `json:"masterId,omitempty" dynamodbav:"masterId,omitempty"`
	Title       *string           `json:"title,omitempty" dynamodbav:"title,omitempty"`
	EventDay    *string           `json:"eventDay,omitempty" dynamodbav:"eventDay,omitempty"`
	Day         *string           `json:"day,omitempty" dynamodbav:"day,omitempty"`
	Organizer   *string           `json:"organizer,omitempty" dynamodbav:"organizer,omitempty"`
	Commit      *string           `json:"commit,omitempty" dynamodbav:"commit,omitempty"`
	Beginner    *string           `json:"beginner,omitempty" dynamodbav:"beginner,omitempty"`
	Message     *string           `json:"message,omitempty" dynamodbav:"message,omitempty"`
	SlackUrl    *string           `json:"slackUrl,omitempty" dynamodbav:"slackUrl,omitempty"`
	TotalMember *string           `json:"totalMember,omitempty" dynamodbav:"totalMember,omitempty"`
	Position    *string           `json:"position,omitempty" dynamodbav:"position,omitempty"`
	Reword      *string           `json:"reword,omitempty" dynamodbav:"reword,omitempty"`
	Members     []RecruitsMembers `json:"members,omitempty" dynamodbav:"members,omitempty"`
	Created     *string           `json:"created,omitempty" dynamodbav:"created,omitempty"`
	Updated     *string           `json:"updated,omitempty" dynamodbav:"updated,omitempty"`
}
type AllGetType []RecruitAllGetResponse

func (s *Server) RecruitAllGet(w http.ResponseWriter, r *http.Request) {
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

	var resRecruit = make(AllGetType, 0)
	dynamodbattribute.UnmarshalListOfMaps(result.Items, &resRecruit)
	sort.Sort(resRecruit)
	j, _ := json.Marshal(resRecruit)
	w.Write(j)

	// 取得値のログ
	fmt.Println(string(j))
}

// ==================== Get ====================
type RecruitGetResponse struct {
	Id          *int              `json:"id,omitempty" dynamodbav:"id,omitempty"`
	MasterId    *string           `json:"masterId,omitempty" dynamodbav:"masterId,omitempty"`
	Title       *string           `json:"title,omitempty" dynamodbav:"title,omitempty"`
	EventDay    *string           `json:"eventDay,omitempty" dynamodbav:"eventDay,omitempty"`
	Day         *string           `json:"day,omitempty" dynamodbav:"day,omitempty"`
	Organizer   *string           `json:"organizer,omitempty" dynamodbav:"organizer,omitempty"`
	Commit      *string           `json:"commit,omitempty" dynamodbav:"commit,omitempty"`
	Beginner    *string           `json:"beginner,omitempty" dynamodbav:"beginner,omitempty"`
	Message     *string           `json:"message,omitempty" dynamodbav:"message,omitempty"`
	SlackUrl    *string           `json:"slackUrl,omitempty" dynamodbav:"slackUrl,omitempty"`
	TotalMember *string           `json:"totalMember,omitempty" dynamodbav:"totalMember,omitempty"`
	Position    *string           `json:"position,omitempty" dynamodbav:"position,omitempty"`
	Reword      *string           `json:"reword,omitempty" dynamodbav:"reword,omitempty"`
	Members     []RecruitsMembers `json:"members,omitempty" dynamodbav:"members,omitempty"`
	Created     *string           `json:"created,omitempty" dynamodbav:"created,omitempty"`
	Updated     *string           `json:"updated,omitempty" dynamodbav:"updated,omitempty"`
}

func (s *Server) RecruitGet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	tableName := "Recruits"
	active := true
	param := &dynamodb.ScanInput{
		TableName:        aws.String(tableName),
		FilterExpression: aws.String("#I = :id AND #A = :active"),
		ExpressionAttributeNames: map[string]*string{
			"#I": aws.String("id"),
			"#A": aws.String("isActive"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":id": {
				N: aws.String(vars["id"]),
			},
			":active": {
				BOOL: &active,
			},
		},
	}
	result, _ := s.db.Scan(param)

	var resRecruit RecruitGetResponse
	dynamodbattribute.UnmarshalMap(result.Items[0], &resRecruit)
	j, _ := json.Marshal(resRecruit)
	w.Write(j)

	// 取得値のログ
	fmt.Println(string(j))
}

// ==================== Create ====================
type RecruitsCreateRequest struct {
	Id          *int              `json:"id,omitempty" dynamodbav:"id,omitempty"`
	MasterId    *string           `json:"masterId,omitempty" dynamodbav:"masterId,omitempty"`
	Title       *string           `json:"title,omitempty" dynamodbav:"title,omitempty"`
	EventDay    *string           `json:"eventDay,omitempty" dynamodbav:"eventDay,omitempty"`
	Day         *string           `json:"day,omitempty" dynamodbav:"day,omitempty"`
	Organizer   *string           `json:"organizer,omitempty" dynamodbav:"organizer,omitempty"`
	Commit      *string           `json:"commit,omitempty" dynamodbav:"commit,omitempty"`
	Beginner    *string           `json:"beginner,omitempty" dynamodbav:"beginner,omitempty"`
	Message     *string           `json:"message,omitempty" dynamodbav:"message,omitempty"`
	SlackUrl    *string           `json:"slackUrl,omitempty" dynamodbav:"slackUrl,omitempty"`
	TotalMember *string           `json:"totalMember,omitempty" dynamodbav:"totalMember,omitempty"`
	Position    *string           `json:"position,omitempty" dynamodbav:"position,omitempty"`
	Reword      *string           `json:"reword,omitempty" dynamodbav:"reword,omitempty"`
	Members     []RecruitsMembers `json:"members" dynamodbav:"members"`
	Created     *string           `json:"created,omitempty" dynamodbav:"created,omitempty"`
	Updated     *string           `json:"updated,omitempty" dynamodbav:"updated,omitempty"`
	IsActive    bool              `json:"isActive" dynamodbav:"isActive"`
}

func (s *Server) RecruitCreate(w http.ResponseWriter, r *http.Request) {
	nowTime := time.Now().UTC().In(
		time.FixedZone("Asia/Tokyo", 9*60*60),
	).Format("2006-01-02 15:04")

	tableName := "Recruits"
	// 連番の取得
	id := s.Increment(tableName)

	var reqRecruit RecruitsCreateRequest
	json.Unmarshal(StreamToByte(r.Body), &reqRecruit)

	reqRecruit.Id = &id
	reqRecruit.Created = &nowTime
	reqRecruit.Updated = &nowTime
	reqRecruit.IsActive = true
	reqRecruit.Members = append([]RecruitsMembers{}, RecruitsMembers{
		Uid:      reqRecruit.MasterId,
		Position: reqRecruit.Position,
	})

	av, err := dynamodbattribute.MarshalMap(reqRecruit)

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

	j, _ := json.Marshal(reqRecruit)

	// 作成値のログ
	fmt.Println(string(j))
}

// ==================== Member Add ====================
type MemberAddRequest struct {
	Uid      *string `json:"uid,omitempty" dynamodbav:"uid,omitempty"`
	Position *string `json:"position,omitempty" dynamodbav:"position,omitempty"`
}

func (s *Server) MemberAdd(w http.ResponseWriter, r *http.Request) {
	nowTime := time.Now().UTC().In(
		time.FixedZone("Asia/Tokyo", 9*60*60),
	).Format("2006-01-02 15:04")

	vars := mux.Vars(r)

	var reqMember MemberAddRequest
	json.Unmarshal(StreamToByte(r.Body), &reqMember)

	addMap := map[string]*dynamodb.AttributeValue{
		"uid": {
			S: aws.String(*reqMember.Uid),
		},
		"position": {
			S: aws.String(*reqMember.Position),
		},
	}
	addList := []*dynamodb.AttributeValue{
		{
			M: addMap,
		},
	}

	tableName := "Recruits"
	param := &dynamodb.UpdateItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				N: aws.String(vars["id"]),
			},
		},
		UpdateExpression: aws.String(
			"set #members = list_append(#members, :addend), #updated = :updated",
		),
		ExpressionAttributeNames: map[string]*string{
			"#members": aws.String("members"),
			"#updated": aws.String("updated"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":addend": {
				L: addList,
			},
			":updated": {
				S: aws.String(nowTime),
			},
		},
	}
	_, err := s.db.UpdateItem(param)

	if err != nil {
		fmt.Println("Got error calling PutItem:")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	j, _ := json.Marshal(reqMember)
	// 追加メンバーのログ
	fmt.Println(string(j))

	// 募集者にメール送信
	recruitMail := s.RecruitMailInfo(vars["id"], *reqMember.Position)
	s.MailSend(recruitMail)
	
	// 参加者にメール送信
	joinMail := s.JoinMailInfo(*reqMember.Uid, *reqMember.Position, vars["id"])
	s.MailSend(joinMail)
}

type MailInfo struct {
	Sender    string // 送信元メールアドレス
	Recipient string // 宛先メールアドレス
	// Specify a configuration set. If you do not want to use a configuration
	// set, comment out the following constant and the
	// ConfigurationSetName: aws.String(ConfigurationSet) argument below
	// ConfigurationSet string  //
	Subject  string // 件名
	HtmlBody string // メールのhtml本文
	TextBody string // 受信者の電子メール本文。
	CharSet  string // 文字のエンコード
}

func NewMailInfo(sender, name, charSet string) *MailInfo {
	return &MailInfo{
		Sender:  name + "<" + sender + ">",
		CharSet: charSet,
	}
}

type GetRecruit struct {
	MasterId *string           `dynamodbav:"masterId,omitempty"`
	Title    *string           `dynamodbav:"title,omitempty"`
	EventDay *string           `dynamodbav:"eventDay,omitempty"`
	Day      *string           `dynamodbav:"day,omitempty"`
	SlackUrl *string           `dynamodbav:"slackUrl,omitempty"`
	Members  []RecruitsMembers `dynamodbav:"members,omitempty"`
}
type GetUser struct {
	Name  *string `dynamodbav:"name,omitempty"`
	Email *string `dynamodbav:"email,omitempty"`
}

func (s *Server) RecruitMailInfo(id, position string) *MailInfo {
	mailInfo := *NewMailInfo("info@raityupiyo.dev", "GuildHack", "UTF-8")

	result, _ := s.db.Scan(&dynamodb.ScanInput{
		TableName:                 aws.String("Recruits"),
		FilterExpression:          aws.String("#I = :id"),
		ExpressionAttributeNames:  map[string]*string{"#I": aws.String("id")},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{":id": {N: aws.String(id)}},
	})
	var getRecruit GetRecruit
	dynamodbattribute.UnmarshalMap(result.Items[0], &getRecruit)

	result, _ = s.db.Scan(&dynamodb.ScanInput{
		TableName:                 aws.String("EndUsers"),
		FilterExpression:          aws.String("#u = :uid"),
		ExpressionAttributeNames:  map[string]*string{"#u": aws.String("uid")},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{":uid": {S: aws.String(*getRecruit.MasterId)}},
	})
	var getUser GetUser
	dynamodbattribute.UnmarshalMap(result.Items[0], &getUser)

	positionList := map[string]string{
		"frontend": "フロントエンド",
		"backend":  "バックエンド",
		"infra":    "インフラ",
	}

	mailInfo.Recipient = *getUser.Email
	mailInfo.Subject = "【GuildHack】募集中ボードに参加メンバー追加の通知"
	mailInfo.HtmlBody = "<p>" + *getUser.Name + "さん、こんにちは！ GuildHack運営事務局です。</p>" +
		"<p>タイトル : " + *getRecruit.Title + "のボードに" + positionList[position] + "で" + strconv.Itoa(len(getRecruit.Members)-1) + "人目のメンバーが参加しました。</p>" +
		"<p>以下のURLをクリックし、確認してください。</p>" +
		"<p><a href='https://raityupiyo.dev/quest_bord/" + id + "'>https://raityupiyo.dev/quest_bord/" + id + "</a></p>" +
		"<br>" +
		"<p>※イベント参加時のトラブルの責任は一切おいかねますので、ご了承ください。</p>" +
		"<p>※連絡のない当日不参加が繰り返される場合、退会とさせていただくことがありますので、ご了承ください。</p>" +
		"<p>※本メールアドレスは送信専用のため、返信できません。</p>" +
		"<p>---------------------------</p>" +
		"<p>GuildHack運営事務局</p>" +
		"<p>Mail : support@raityupiyo.dev</p>" +
		"<p><a href='https://raityupiyo.dev'>https://raityupiyo.dev</a></p>" +
		"<p>---------------------------</p>"
	mailInfo.TextBody = *getUser.Name + "さん、こんにちは！ GuildHack運営事務局です。\n" +
		"タイトル : " + *getRecruit.Title + "のボードに" + positionList[position] + "で" + strconv.Itoa(len(getRecruit.Members)-1) + "/のメンバーが参加しました。\n" +
		"以下のURLをクリックし、確認してください。\n" +
		"https://raityupiyo.dev/quest_bord/" + id + "\n" +
		"\n" +
		"※イベント参加時のトラブルの責任は一切おいかねますので、ご了承ください。\n" +
		"※連絡のない当日不参加が繰り返される場合、退会とさせていただくことがありますので、ご了承ください。\n" +
		"※本メールアドレスは送信専用のため、返信できません。\n" +
		"---------------------------\n" +
		"GuildHack運営事務局\n" +
		"Mail : support@raityupiyo.dev\n" +
		"https://raityupiyo.dev\n" +
		"--------------------------"
	return &mailInfo
}

func (s *Server) JoinMailInfo(uid, position, id string) *MailInfo {
	mailInfo := *NewMailInfo("info@raityupiyo.dev", "GuildHack", "UTF-8")

	result, _ := s.db.Scan(&dynamodb.ScanInput{
		TableName:                 aws.String("Recruits"),
		FilterExpression:          aws.String("#I = :id"),
		ExpressionAttributeNames:  map[string]*string{"#I": aws.String("id")},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{":id": {N: aws.String(id)}},
	})
	var getRecruit GetRecruit
	dynamodbattribute.UnmarshalMap(result.Items[0], &getRecruit)

	result, _ = s.db.Scan(&dynamodb.ScanInput{
		TableName:                 aws.String("EndUsers"),
		FilterExpression:          aws.String("#u = :uid"),
		ExpressionAttributeNames:  map[string]*string{"#u": aws.String("uid")},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{":uid": {S: aws.String(uid)}},
	})
	var getUser GetUser
	dynamodbattribute.UnmarshalMap(result.Items[0], &getUser)

	positionList := map[string]string{
		"frontend": "フロントエンド",
		"backend":  "バックエンド",
		"infra":    "インフラ",
	}

	mailInfo.Recipient = *getUser.Email
	mailInfo.Subject = "【GuildHack】参加完了の通知"
	mailInfo.HtmlBody = "<p>" + *getUser.Name + "さん、こんにちは！ GuildHack運営事務局です。</p>" +
		"<p>タイトル : " + *getRecruit.Title + "（" + *getRecruit.EventDay + "からの" + *getRecruit.Day + "日間）に" + positionList[position] + "として参加が確定しました。</p>" +
		"<p>以下のURLをクリックし、確認してください。</p>" +
		"<p><a href='https://raityupiyo.dev/quest_bord/" + id + "'>https://raityupiyo.dev/quest_bord/" + id + "</a></p>" +
		"<br>" +
		"<p>コミュニケーションツールへの招待は以下のURLになります。</p>" +
		"<p><a href='" + *getRecruit.SlackUrl + "'>" + *getRecruit.SlackUrl + "</a></p>" +
		"<br>" +
		"<p>※イベント参加時のトラブルの責任は一切おいかねますので、ご了承ください。</p>" +
		"<p>※連絡のない当日不参加が繰り返される場合、退会とさせていただくことがありますので、ご了承ください。</p>" +
		"<p>※本メールアドレスは送信専用のため、返信できません。</p>" +
		"<p>---------------------------</p>" +
		"<p>GuildHack運営事務局</p>" +
		"<p>Mail : support@raityupiyo.dev</p>" +
		"<p><a href='https://raityupiyo.dev'>https://raityupiyo.dev</a></p>" +
		"<p>---------------------------</p>"
	mailInfo.TextBody = *getUser.Name + "さん、こんにちは！ GuildHack運営事務局です。\n" +
		"タイトル : " + *getRecruit.Title + "（" + *getRecruit.EventDay + "からの" + *getRecruit.Day + "日間）への参加が確定しました。\n" +
		"以下のURLをクリックし、確認してください。\n" +
		"https://raityupiyo.dev/quest_bord/" + id + "\n" +
		"\n" +
		"コミュニケーションツールへの招待は以下のURLになります。" +
		*getRecruit.SlackUrl + "\n" +
		"\n" +
		"※イベント参加時のトラブルの責任は一切おいかねますので、ご了承ください。\n" +
		"※連絡のない当日不参加が繰り返される場合、退会とさせていただくことがありますので、ご了承ください。\n" +
		"※本メールアドレスは送信専用のため、返信できません。\n" +
		"---------------------------\n" +
		"GuildHack運営事務局\n" +
		"Mail : support@raityupiyo.dev\n" +
		"https://raityupiyo.dev\n" +
		"--------------------------"
	return &mailInfo
}

func (s *Server) MailSend(info *MailInfo) {
	input := &ses.SendEmailInput{
		Destination: &ses.Destination{
			CcAddresses: []*string{},
			ToAddresses: []*string{
				aws.String(info.Recipient),
			},
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Html: &ses.Content{
					Charset: aws.String(info.CharSet),
					Data:    aws.String(info.HtmlBody),
				},
				Text: &ses.Content{
					Charset: aws.String(info.CharSet),
					Data:    aws.String(info.TextBody),
				},
			},
			Subject: &ses.Content{
				Charset: aws.String(info.CharSet),
				Data:    aws.String(info.Subject),
			},
		},
		Source: aws.String(info.Sender),
		// Comment or remove the following line if you are not using a configuration set
		//ConfigurationSetName: aws.String(info.ConfigurationSet),
	}
	result, err := s.ses.SendEmail(input)

	// Display error messages if they occur.
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case ses.ErrCodeMessageRejected:
				fmt.Println(ses.ErrCodeMessageRejected, aerr.Error())
			case ses.ErrCodeMailFromDomainNotVerifiedException:
				fmt.Println(ses.ErrCodeMailFromDomainNotVerifiedException, aerr.Error())
			case ses.ErrCodeConfigurationSetDoesNotExistException:
				fmt.Println(ses.ErrCodeConfigurationSetDoesNotExistException, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return
	}

	// メールのログ
	fmt.Println(result)
}
