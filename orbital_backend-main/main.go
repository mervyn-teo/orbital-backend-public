package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

// empty profile
var nullProf = profile{"", "", 0, "", ""}

// global use database
var dataBase *sql.DB

// formatting
type profile struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
	Bio  string `json:"bio"`
	Pfp  string `json:"pfp"`
}

type auth struct {
	Email string `json:"email"`
	Pwd   string `json:"pwd"`
}

type tag struct {
	Id  string `json:"id"`
	Tag string `json:"tag"`
}

type idPair struct {
	IdFrom string `json:"id_from"`
	IdTo   string `json:"id_to"`
}

type id struct {
	Id string `json:"id"`
}

type geog struct {
	Id   string  `json:"id"`
	Lat  float64 `json:"lat"`
	Long float64 `json:"long"`
}

type response struct {
	Err_msg string        `json:"err_msg"`
	Body    []interface{} `json:"body"`
}

type message struct {
	MsgId    int       `json:"msg_id"`
	IdFrom   string    `json:"id_from"`
	IdTo     string    `json:"id_to"`
	Msg      string    `json:"msg"`
	TimeSent time.Time `json:"time_sent"`
}

type Server struct {
	conns map[*websocket.Conn]bool
}

type event struct {
	EventId     string    `json:"event_id"`
	UserIds     []int64   `json:"user_id"`
	Size        int       `json:"size"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	DateTime    time.Time `json:"date_time"`
	Owner       string    `json:"owner"`
}

type eventIdPair struct {
	EventId string `json:"event_id"`
	UserId  string `json:"user_id"`
}

type addEvent struct {
	UserId  int64  `json:"id"`
	EventId string `json:"event_id"`
}

func newResponse(err_msg string, body []interface{}) response {
	var newResponse response
	newResponse.Err_msg = err_msg
	newResponse.Body = body
	return newResponse
}

// intialise things
func init() {
	// initalise database
	connStr := "user=postgres dbname=orbital password=thisisalongpass host=13.231.75.235 sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	// make database global
	dataBase = db
}

// grab all profiles
func getProfiles(context *gin.Context) {

	profiles := []profile{}

	rows, err := dataBase.Query(`SELECT * FROM profile;`)
	if err != nil {
		log.Fatal(err)
	}

	for rows.Next() {
		prof := new(profile)
		rows.Scan(&prof.ID, &prof.Name, &prof.Age, &prof.Bio, &prof.Pfp)
		profiles = append(profiles, *prof)
	}
	context.IndentedJSON(http.StatusOK, newResponse("ok", []interface{}{profiles}))
}

func addProfile(context *gin.Context) {
	var newProfile profile

	if err := context.BindJSON(&newProfile); err != nil {
		println(err.Error())
		return
	}

	sqlStatement := `
		INSERT INTO profile (id, name, age, bio, pfp)
		VALUES ($1, $2, $3, $4, $5)`
	_, err := dataBase.Exec(sqlStatement, newProfile.ID, newProfile.Name, newProfile.Age, newProfile.Bio, newProfile.Pfp)
	if err != nil {
		panic(err)
		context.IndentedJSON(http.StatusBadRequest, newResponse(err.Error(), []interface{}{newProfile}))
	} else {
		context.IndentedJSON(http.StatusCreated, newResponse("ok", []interface{}{newProfile}))
	}
}

func emailExists(email string, context *gin.Context) bool {
	var ret int
	sqlStatement := `SELECT count(*) FROM auth WHERE email=$1 LIMIT 1`
	err := dataBase.QueryRow(sqlStatement, email).Scan(&ret)
	if err != nil {
		panic(err)
	} else {
		fmt.Println("returned:", ret)
		return ret > 0
	}
}

// generate a random string of length 10
func generateSalt() string {
	length := 10
	charset := "abcdefghijklmnopqrstuvwxyz"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func register(context *gin.Context) {
	var newAuth auth

	if err := context.BindJSON(&newAuth); err != nil {
		return
	}

	if emailExists(newAuth.Email, context) {
		context.IndentedJSON(http.StatusConflict, "Email exists!")
	} else {
		salt := generateSalt()
		salted := newAuth.Pwd + salt
		hash := md5.Sum([]byte(salted))
		pwd := hex.EncodeToString(hash[:])
		var id int

		//create new row in auth table
		newAuthStatement := `
			INSERT INTO auth (email, salt, pwd)
			VALUES ($1, $2, $3)
		`
		_, err := dataBase.Exec(newAuthStatement, newAuth.Email, salt, pwd)

		if err != nil {
			panic(err)
		} else {
			err2 := dataBase.QueryRow("SELECT id FROM auth WHERE email=$1", newAuth.Email).Scan(&id)
			if err2 != nil {
				panic(err2)
			}
			context.IndentedJSON(http.StatusCreated, newResponse("ok", []interface{}{newAuth, id}))
		}
	}
}

// TODO: add authorisation to this
func editProfile(context *gin.Context) {
	var newProfile profile

	if err := context.BindJSON(&newProfile); err != nil {
		return
	}

	editProfileStatement := `
		UPDATE profile
		SET name=$1, 
			age=$2, 
			bio=$3, 
			pfp=$4
		WHERE id=$5;
	`
	_, err := dataBase.Exec(editProfileStatement, newProfile.Name, newProfile.Age, newProfile.Bio, newProfile.Pfp, newProfile.ID)

	if err != nil {
		panic(err)
		context.IndentedJSON(http.StatusBadRequest, newResponse(err.Error(), []interface{}{newProfile}))
	} else {
		context.IndentedJSON(http.StatusOK, newResponse("ok", []interface{}{newProfile}))
	}
}

func login(context *gin.Context) {
	var loginAuth auth
	var recordedSalt string
	var recordedPwd string
	var recordedId int

	if err := context.BindJSON(&loginAuth); err != nil {
		return
	}

	newLoginStatement := `
		SELECT salt, pwd, id FROM auth WHERE email=$1
	`
	err := dataBase.QueryRow(newLoginStatement, loginAuth.Email).Scan(&recordedSalt, &recordedPwd, &recordedId)

	if err != nil {
		if err == sql.ErrNoRows {
			// email not found
			context.IndentedJSON(http.StatusNotFound, newResponse("email not found", []interface{}{loginAuth}))
		} else {
			panic(err)
		}
	} else {
		salted := loginAuth.Pwd + recordedSalt
		hash := md5.Sum([]byte(salted))
		userInputPassword := hex.EncodeToString(hash[:])

		if userInputPassword == recordedPwd {
			// pwd correct
			returnProfile := retriveProfile(recordedId)

			if returnProfile == nullProf {
				context.IndentedJSON(http.StatusFound, newResponse("error while retriving profile", []interface{}{returnProfile}))
			}

			context.IndentedJSON(http.StatusFound, newResponse("ok", []interface{}{returnProfile}))
		} else {
			//pwd not correct
			context.IndentedJSON(http.StatusExpectationFailed, newResponse("pwd incorrect", []interface{}{loginAuth}))
		}
	}
}

func retriveProfile(id int) profile {
	var profile profile
	err := dataBase.QueryRow(`SELECT age, bio, name, pfp FROM profile WHERE id=$1;`, id).Scan(&profile.Age, &profile.Bio, &profile.Name, &profile.Pfp)
	profile.ID = strconv.Itoa(id)

	if err != nil {
		log.Fatal(err)
		return nullProf
	}
	return profile
}

func addTag(context *gin.Context) {
	var myTag tag
	var count int

	if err := context.BindJSON(&myTag); err != nil {
		return
	}

	// check if theres more than number allowed
	checkTagStatement := `
		SELECT COUNT(tag)
		FROM tags
		WHERE id = $1
	`
	err := dataBase.QueryRow(checkTagStatement, myTag.Id).Scan(&count)

	if err != nil {
		log.Fatal(err)
		context.IndentedJSON(http.StatusOK, newResponse(err.Error(), []interface{}{myTag}))
	}

	if count < 6 {
		// check if theres more than number allowed
		checkRepeatedTagStatement := `
			SELECT COUNT(tag)
			FROM tags
			WHERE id = $1 AND tag = $2
		`
		err = dataBase.QueryRow(checkRepeatedTagStatement, myTag.Id, myTag.Tag).Scan(&count)

		if err != nil {
			log.Fatal(err)
			context.IndentedJSON(http.StatusOK, newResponse(err.Error(), []interface{}{myTag}))
		}

		if count == 0 {
			// adding the tag
			addTagStatement := `
				INSERT INTO tags (id, tag)
				VALUES ($1, $2)
			`
			_, err = dataBase.Exec(addTagStatement, myTag.Id, myTag.Tag)

			if err != nil {
				panic(err)
			} else {
				context.IndentedJSON(http.StatusOK, newResponse("ok", []interface{}{myTag}))
			}
		} else {
			context.IndentedJSON(http.StatusOK, newResponse("Tag existed!", []interface{}{myTag}))
		}
	} else {
		context.IndentedJSON(http.StatusOK, newResponse("Maximum amount of tags reached!", []interface{}{myTag}))
	}

}

func queryTag(context *gin.Context) {
	var id id
	tags := []tag{}

	if err := context.BindJSON(&id); err != nil {
		return
	}

	rows, err := dataBase.Query(`SELECT * FROM tags WHERE id=$1;`, id.Id)
	if err != nil {
		log.Fatal(err)
		context.IndentedJSON(http.StatusOK, newResponse(err.Error(), []interface{}{tags}))
	}

	for rows.Next() {
		tag := new(tag)
		rows.Scan(&tag.Id, &tag.Tag)
		tags = append(tags, *tag)
	}
	context.IndentedJSON(http.StatusOK, newResponse("ok", []interface{}{tags}))
}

func deleteTag(context *gin.Context) {
	var tag tag

	if err := context.BindJSON(&tag); err != nil {
		return
	}

	_, err := dataBase.Exec(`DELETE FROM tags WHERE id=$1 AND tag=$2`, tag.Id, tag.Tag)
	if err != nil {
		log.Fatal(err)
		context.IndentedJSON(http.StatusOK, newResponse(err.Error(), []interface{}{tag}))
	}

	context.IndentedJSON(http.StatusOK, newResponse("ok", []interface{}{tag}))
}

func updateGeog(context *gin.Context) {
	var geog geog

	if err := context.BindJSON(&geog); err != nil {
		return
	}

	query := `
		INSERT INTO geog (id, point, time) 
		VALUES ($1, ST_SetSRID(ST_MakePoint($2, $3),4326), current_timestamp)
		ON CONFLICT (id) 
		DO UPDATE SET point = ST_SetSRID(ST_MakePoint($2, $3),4326), time = current_timestamp;
	`

	_, err := dataBase.Exec(query, geog.Id, geog.Long, geog.Lat)

	if err != nil {
		context.IndentedJSON(http.StatusOK, newResponse(err.Error(), []interface{}{geog}))
		log.Fatal(err)
	}

	distQuery := fmt.Sprintf(
		`DO $$
		DECLARE
			r RECORD;
		BEGIN
			FOR r IN
				SELECT * FROM geog
				WHERE ST_DWithin(ST_MAKEPOINT(%f, %f), geog.point, 50, false)
				AND EXTRACT(EPOCH FROM (current_timestamp - geog.time)) < 30 AND id != %s 
			LOOP
				IF (SELECT COUNT(*) FROM encounter WHERE id = %s AND oppid = r.id ) < 1 THEN
					INSERT INTO encounter (id, oppid, count)
					VALUES (%s, r.id, 1);
				ELSE
					UPDATE encounter SET count = count + 1 WHERE id = %s AND oppid = r.id;
				END IF;
			END LOOP;
		END; $$;`, geog.Long, geog.Lat, geog.Id, geog.Id, geog.Id, geog.Id) //all entris within 50 m && 30s

	_, err = dataBase.Exec(distQuery)

	if err != nil {
		log.Fatal(err)
	}

	context.IndentedJSON(http.StatusOK, newResponse("ok", []interface{}{geog}))
}

func metNumber(context *gin.Context) {
	var id id
	var ret int
	if err := context.BindJSON(&id); err != nil {
		return
	}

	err := dataBase.QueryRow(`SELECT COALESCE(SUM(count), 0) FROM encounter WHERE id=$1 OR oppid=$1`, id.Id).Scan(&ret)
	if err != nil {
		log.Fatal(err)
		context.IndentedJSON(http.StatusOK, newResponse(err.Error(), []interface{}{id}))
	}

	context.IndentedJSON(http.StatusOK, newResponse("ok", []interface{}{ret}))
}

func matches(context *gin.Context) {
	var id id
	matchedTags := make(map[string][]string)
	matchingTagAmount := make(map[string][]float64)
	tags := []tag{}

	if err := context.BindJSON(&id); err != nil {
		return
	}

	// get all tags
	rows, err := dataBase.Query(`SELECT * FROM tags WHERE id=$1;`, id.Id)
	if err != nil {
		log.Fatal(err)
		context.IndentedJSON(http.StatusOK, newResponse(err.Error(), []interface{}{tags}))
	}

	for rows.Next() {
		tag := new(tag)
		rows.Scan(&tag.Id, &tag.Tag)
		tags = append(tags, *tag)
	}

	// get all ids that consist tags
	for i := 0; i < len(tags); i++ {

		rows, err = dataBase.Query(`
			SELECT * FROM tags WHERE tag=$1 AND 
				((SELECT COUNT(*) FROM interested WHERE id_from=$2 AND id_to=tags.id) < 1) AND 
				((SELECT COUNT(*) FROM not_interested WHERE id_from=$2 AND id_to=tags.id) < 1);
		`, tags[i].Tag, id.Id)
		if err != nil {
			log.Fatal(err)
			context.IndentedJSON(http.StatusOK, newResponse(err.Error(), []interface{}{tags}))
		}

		for rows.Next() {
			temp := new(tag)
			rows.Scan(&temp.Id, &temp.Tag)
			if matchedTags[temp.Id] == nil {
				matchedTags[temp.Id] = []string{tags[i].Tag}
			} else {
				matchedTags[temp.Id] = append(matchedTags[temp.Id], tags[i].Tag)
			}
		}
	}

	// add sizes
	for id, items := range matchedTags {
		matchingTagAmount[id] = make([]float64, 2)
		if items != nil {
			matchingTagAmount[id][0] = float64(len(items))
		} else {
			matchingTagAmount[id][0] = 0.0
		}

	}

	keys := make([]string, 0, len(matchingTagAmount))

	for key := range matchingTagAmount {
		keys = append(keys, key)
	}

	// query for total number of met
	var totalMet int
	err = dataBase.QueryRow(`SELECT COALESCE(SUM(count), 0) FROM encounter WHERE id=$1 OR oppid=$1`, id.Id).Scan(&totalMet)
	if err != nil {
		fmt.Println(1)
		log.Fatal(err)
	}

	// calculate meet percentage
	for tempId, _ := range matchingTagAmount {
		if tempId != "" {
			var tempMet int
			err = dataBase.QueryRow(`SELECT COALESCE(SUM(count), 0) FROM encounter WHERE ((id=$1 AND oppid=$2) OR (id=$2 AND oppid=$1))`, id.Id, tempId).Scan(&tempMet)
			if err != nil {
				fmt.Println(2)
				log.Fatal(err)
				matchingTagAmount[tempId][1] = 0.0
			} else {
				matchingTagAmount[tempId][1] = float64(tempMet) / float64(totalMet)
			}
		} else {
			context.IndentedJSON(http.StatusOK, newResponse("no match found", []interface{}{id}))
			return
		}
	}

	// sort the keys based on value
	sort.SliceStable(keys, func(i, j int) bool {
		return matchingTagAmount[keys[i]][0]*0.5+matchingTagAmount[keys[i]][1]*0.5 > matchingTagAmount[keys[j]][0]*0.5+matchingTagAmount[keys[j]][1]*0.5
	})

	var returnProf []profile
	// return top 10 profiles
	for i := 0; i < 10 && i < len(keys); i++ {
		rows, err := dataBase.Query(`SELECT * FROM profile WHERE id=$1;`, keys[i])
		if err != nil {
			log.Fatal(err)
		}

		for rows.Next() {
			prof := new(profile)
			rows.Scan(&prof.ID, &prof.Name, &prof.Age, &prof.Bio, &prof.Pfp)
			returnProf = append(returnProf, *prof)
		}
	}

	context.IndentedJSON(http.StatusOK, newResponse("ok", []interface{}{returnProf}))
}

func addInterest(context *gin.Context) {
	var newInterest idPair

	if err := context.BindJSON(&newInterest); err != nil {
		return
	}

	sqlStatement := fmt.Sprintf(`
	DO $$
		BEGIN
			IF (SELECT COUNT(*) FROM interested WHERE id_from = %s AND id_to = %s ) < 1 THEN
				INSERT INTO interested (id_from, id_to) VALUES (%s, %s);
			END IF;
		END; 
	$$;
	`, newInterest.IdFrom, newInterest.IdTo, newInterest.IdFrom, newInterest.IdTo)

	_, err := dataBase.Exec(sqlStatement)

	if err != nil {
		log.Fatal(err)
		context.IndentedJSON(http.StatusOK, newResponse(err.Error(), []interface{}{newInterest}))
	}

	context.IndentedJSON(http.StatusOK, newResponse("ok", []interface{}{newInterest}))
}

func addNotInterest(context *gin.Context) {
	var newInterest idPair

	if err := context.BindJSON(&newInterest); err != nil {
		return
	}

	sqlStatement := fmt.Sprintf(`
	DO $$
		BEGIN
			IF (SELECT COUNT(*) FROM not_interested WHERE id_from = %s AND id_to = %s ) < 1 THEN
				INSERT INTO not_interested (id_from, id_to) VALUES (%s, %s);
			END IF;
		END; 
	$$;
	`, newInterest.IdFrom, newInterest.IdTo, newInterest.IdFrom, newInterest.IdTo)

	_, err := dataBase.Exec(sqlStatement)

	if err != nil {
		log.Fatal(err)
		context.IndentedJSON(http.StatusOK, newResponse(err.Error(), []interface{}{newInterest}))
	}

	context.IndentedJSON(http.StatusOK, newResponse("ok", []interface{}{newInterest}))
}

func getMessage(context *gin.Context) {
	var newIdPair message

	if err := context.BindJSON(&newIdPair); err != nil {
		return
	}

	sqlstatement := `
	SELECT * FROM messaage 
	WHERE ((id_from = $1 AND id_to = $2) 
	OR (id_to = $1 AND id_from = $2)) 
	AND (time_sent < $3) 
	ORDER BY time_sent DESC
	LIMIT 10;`

	rows, err := dataBase.Query(sqlstatement, newIdPair.IdFrom, newIdPair.IdTo, newIdPair.TimeSent)

	if err != nil {
		log.Fatal(err)
		context.IndentedJSON(http.StatusOK, newResponse(err.Error(), []interface{}{newIdPair}))
	}

	var returnMsg []message
	for rows.Next() {
		temp := new(message)
		rows.Scan(&temp.MsgId, &temp.IdFrom, &temp.IdTo, &temp.Msg, &temp.TimeSent)
		returnMsg = append(returnMsg, *temp)
	}

	context.IndentedJSON(http.StatusOK, newResponse("ok", []interface{}{returnMsg}))
}

func sendMessage(context *gin.Context) {
	var newMessage message

	if err := context.BindJSON(&newMessage); err != nil {
		return
	}

	sqlstatement := `
	INSERT INTO messaage (id_from, id_to, msg, time_sent) 
	VALUES ($1, $2, $3, current_timestamp) 
	RETURNING * ;`

	var temp message
	err := dataBase.QueryRow(sqlstatement, newMessage.IdFrom, newMessage.IdTo, newMessage.Msg).Scan(&temp.MsgId, &temp.IdFrom, &temp.IdTo, &temp.Msg, &temp.TimeSent)

	if err != nil {
		log.Fatal(err)
		context.IndentedJSON(http.StatusOK, newResponse(err.Error(), []interface{}{newMessage}))
	}

	context.IndentedJSON(http.StatusOK, newResponse("ok", []interface{}{temp}))

}

func getChat(context *gin.Context) {
	var newId id

	if err := context.BindJSON(&newId); err != nil {
		return
	}

	sqlstatement := `
		SELECT id, name, age, bio, pfp
		FROM profile
		JOIN interested AS outer_interest
		ON profile.id = outer_interest.id_to
		WHERE outer_interest.id_from = $1
		AND (
			SELECT COUNT(*)
			FROM interested AS inner_interest
			WHERE inner_interest.id_from = outer_interest.id_to
			AND inner_interest.id_to = $1 
		) > 0
		AND outer_interest.id_from != outer_interest.id_to;
	`

	rows, err := dataBase.Query(sqlstatement, newId.Id)

	var ret []profile
	for rows.Next() {
		temp := new(profile)
		rows.Scan(&temp.ID, &temp.Name, &temp.Age, &temp.Bio, &temp.Pfp)
		ret = append(ret, *temp)
	}

	if err != nil {
		log.Fatal(err)
		context.IndentedJSON(http.StatusOK, newResponse(err.Error(), []interface{}{newId}))
	}

	context.IndentedJSON(http.StatusOK, newResponse("ok", []interface{}{ret}))
}

func setEvent(context *gin.Context) {
	var newEvent event

	if err := context.BindJSON(&newEvent); err != nil {
		fmt.Print(err)
		return
	}

	statement := `
		INSERT INTO events (user_id, size, name, description, datetime, owner)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING * ;
	`

	var temp event
	err := dataBase.QueryRow(
		statement,
		pq.Array(newEvent.UserIds),
		newEvent.Size,
		newEvent.Name,
		newEvent.Description,
		newEvent.DateTime,
		newEvent.UserIds[0]).Scan(
		&temp.EventId,
		pq.Array(&temp.UserIds),
		&temp.Size,
		&temp.Description,
		&temp.Name,
		&temp.DateTime,
		&temp.Owner)

	if err != nil {
		log.Fatal(err)
		context.IndentedJSON(http.StatusOK, newResponse(err.Error(), []interface{}{newEvent}))
	}

	context.IndentedJSON(http.StatusOK, newResponse("ok", []interface{}{temp}))
}

func editEvent(context *gin.Context) {
	var newEvent event

	if err := context.BindJSON(&newEvent); err != nil {
		fmt.Print(err)
		return
	}

	statement := `
		UPDATE events 
		SET size = $1, name = $2, description = $3, datetime= $4
		WHERE event_id = $5
		RETURNING *;
	`

	var temp event
	err := dataBase.QueryRow(
		statement,
		newEvent.Size,
		newEvent.Name,
		newEvent.Description,
		newEvent.DateTime,
		newEvent.EventId).Scan(
		&temp.EventId,
		pq.Array(&temp.UserIds),
		&temp.Size,
		&temp.Description,
		&temp.Name,
		&temp.DateTime,
		&temp.Owner)

	if err != nil {
		log.Fatal(err)
		context.IndentedJSON(http.StatusOK, newResponse(err.Error(), []interface{}{newEvent}))
	}

	context.IndentedJSON(http.StatusOK, newResponse("ok", []interface{}{temp}))
}

func removeEvent(context *gin.Context) {
	var newEventId id

	if err := context.BindJSON(&newEventId); err != nil {
		fmt.Print(err)
		return
	}

	statement := `
		DELETE from events WHERE event_id = $1;
	`
	_, err := dataBase.Exec(statement, newEventId.Id)

	if err != nil {
		log.Fatal(err)
		context.IndentedJSON(http.StatusOK, newResponse(err.Error(), []interface{}{newEventId}))
	}

	context.IndentedJSON(http.StatusOK, newResponse("ok", []interface{}{newEventId}))
}

func removeIdFromEvent(context *gin.Context) {
	var newEventId eventIdPair

	if err := context.BindJSON(&newEventId); err != nil {
		fmt.Print(err)
		return
	}

	statement := `
		UPDATE events 
		SET user_id = array_remove(user_id, $1)
		WHERE event_id = $2;
	`
	_, err := dataBase.Exec(statement, newEventId.UserId, newEventId.EventId)

	if err != nil {
		log.Fatal(err)
		context.IndentedJSON(http.StatusOK, newResponse(err.Error(), []interface{}{newEventId}))
	}

	context.IndentedJSON(http.StatusOK, newResponse("ok", []interface{}{newEventId}))
}

func getEvent(context *gin.Context) {
	events := []event{}

	rows, err := dataBase.Query(`SELECT * FROM events WHERE datetime > now() AND array_length(user_id, 1) < size;`)

	if err != nil {
		log.Fatal(err)
	}

	for rows.Next() {
		temp := new(event)
		rows.Scan(
			&temp.EventId,
			pq.Array(&temp.UserIds),
			&temp.Size,
			&temp.Description,
			&temp.Name,
			&temp.DateTime,
			&temp.Owner)
		events = append(events, *temp)
	}
	context.IndentedJSON(http.StatusOK, newResponse("ok", []interface{}{events}))
}

func getJoinedEvent(context *gin.Context) {
	var id id
	events := []event{}
	if err := context.BindJSON(&id); err != nil {
		fmt.Print(err)
		return
	}

	rows, err := dataBase.Query(`SELECT * FROM events WHERE datetime > now() AND array_length(user_id, 1) < size AND $1 = ANY(user_id);`, id.Id)

	if err != nil {
		log.Fatal(err)
	}

	for rows.Next() {
		temp := new(event)
		rows.Scan(
			&temp.EventId,
			pq.Array(&temp.UserIds),
			&temp.Size,
			&temp.Description,
			&temp.Name,
			&temp.DateTime,
			&temp.Owner)
		events = append(events, *temp)
	}
	context.IndentedJSON(http.StatusOK, newResponse("ok", []interface{}{events}))
}

func addToEvent(context *gin.Context) {
	var newEvent addEvent

	if err := context.BindJSON(&newEvent); err != nil {
		fmt.Print(err)
		return
	}

	// check for duplicates
	duplicateStatement := `
		SELECT user_id, size FROM events WHERE event_id = $1
	`
	var tempId []int64
	var tempSize int

	err := dataBase.QueryRow(duplicateStatement, newEvent.EventId).Scan(pq.Array(&tempId), &tempSize)

	if err != nil {
		log.Fatal(err)
		context.IndentedJSON(http.StatusOK, newResponse(err.Error(), []interface{}{newEvent}))
	}

	for _, item := range tempId {
		if newEvent.UserId == item {
			context.IndentedJSON(http.StatusOK, newResponse("Already added", []interface{}{newEvent}))
			return
		}
	}

	// check for size
	if len(tempId) >= tempSize {
		context.IndentedJSON(http.StatusOK, newResponse("Event full", []interface{}{newEvent}))
		return
	}

	statement := `
		UPDATE events SET user_id = array_append(user_id, $1) WHERE event_id = $2
		RETURNING * ;
	`

	var temp event
	err = dataBase.QueryRow(statement, newEvent.UserId, newEvent.EventId).Scan(
		&temp.EventId,
		pq.Array(&temp.UserIds),
		&temp.Size,
		&temp.Description,
		&temp.Name,
		&temp.DateTime,
		&temp.Owner)

	if err != nil {
		log.Fatal(err)
		context.IndentedJSON(http.StatusOK, newResponse(err.Error(), []interface{}{newEvent}))
	}

	context.IndentedJSON(http.StatusOK, newResponse("ok", []interface{}{temp}))
}

func addReport(context *gin.Context) {
	var newIdPair idPair

	if err := context.BindJSON(&newIdPair); err != nil {
		fmt.Print(err)
		return
	}

	statement := `
		INSERT INTO reported (from_id, to_id)
		VALUES ($1, $2);
	`

	_, err := dataBase.Exec(statement, newIdPair.IdFrom, newIdPair.IdTo)

	if err != nil {
		panic(err)
		context.IndentedJSON(http.StatusBadRequest, newResponse(err.Error(), []interface{}{newIdPair}))
	} else {
		context.IndentedJSON(http.StatusOK, newResponse("ok", []interface{}{newIdPair}))
	}
}

func removeReport(context *gin.Context) {
	var newIdPair idPair

	if err := context.BindJSON(&newIdPair); err != nil {
		fmt.Print(err)
		return
	}

	statement := `
		DELETE FROM reported WHERE from_id = $1 AND to_id = $2;
	`

	_, err := dataBase.Exec(statement, newIdPair.IdFrom, newIdPair.IdTo)

	if err != nil {
		panic(err)
		context.IndentedJSON(http.StatusBadRequest, newResponse(err.Error(), []interface{}{newIdPair}))
	} else {
		context.IndentedJSON(http.StatusOK, newResponse("ok", []interface{}{newIdPair}))
	}
}

func checkReported(context *gin.Context) {
	var newIdPair idPair
	var retBool bool
	var tempNum int

	if err := context.BindJSON(&newIdPair); err != nil {
		fmt.Print(err)
		return
	}

	statement := `
		SELECT count(*) FROM reported WHERE from_id = $1 AND to_id = $2;
	`

	err := dataBase.QueryRow(statement, newIdPair.IdFrom, newIdPair.IdTo).Scan(&tempNum)

	if err != nil {
		panic(err)
		context.IndentedJSON(http.StatusBadRequest, newResponse(err.Error(), []interface{}{newIdPair}))
	} else {
		if tempNum > 0 {
			retBool = true
		} else {
			retBool = false
		}
		context.IndentedJSON(http.StatusOK, newResponse("ok", []interface{}{retBool}))
	}
}

func main() {
	// HTTP router
	router := gin.Default()
	router.GET("/profiles", getProfiles)
	router.POST("/profiles", addProfile)
	router.POST("/register", register)
	router.POST("/login", login)
	router.PATCH("/profile", editProfile)
	router.PUT("/tags", addTag)
	router.POST("/tags", queryTag)
	router.DELETE("/tags", deleteTag)
	router.PUT("/updateGeog", updateGeog)
	router.POST("/metNumber", metNumber)
	router.POST("/matches", matches)
	router.POST("/addInterest", addInterest)
	router.POST("/addNotInterest", addNotInterest)
	router.POST("/getMessage", getMessage)
	router.POST("/sendMessage", sendMessage)
	router.POST("/getChat", getChat)
	router.POST("/setEvent", setEvent)
	router.PATCH("/editEvent", editEvent)
	router.DELETE("/removeEvent", removeEvent)
	router.GET("/getEvent", getEvent)
	router.POST("/addIdToEvent", addToEvent)
	router.POST("/removeIdFromEvent", removeIdFromEvent)
	router.POST("/getJoinedEvent", getJoinedEvent)
	router.POST("/reported", addReport)
	router.DELETE("/reported", removeReport)
	router.POST("/checkReported", checkReported)

	router.Run(":8080")
}
