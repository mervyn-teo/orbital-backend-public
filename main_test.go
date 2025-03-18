package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Initialize Gin engine and routes for testing
func setupRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	router.GET("/profiles", getProfiles)
	router.POST("/profiles", addProfile)
	router.POST("/register", register)
	router.POST("/login", login)
	router.PUT("/profiles", editProfile)
	router.POST("/tags", addTag)
	router.GET("/tags", queryTag)
	router.DELETE("/tags", deleteTag)
	router.PUT("/geog", updateGeog)
	router.GET("/metNumber", metNumber)
	router.GET("/matches", matches)
	router.POST("/interest", addInterest)
	router.POST("/notInterest", addNotInterest)
	router.GET("/message", getMessage)
	router.POST("/message", sendMessage)
	router.GET("/chat", getChat)

	return router
}

func TestGetProfiles(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/profiles", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLogin(t *testing.T) {
	router := setupRouter()

	auth := auth{
		Email: "testtes123@abc.com",
		Pwd:   "asdfghj",
	}
	jsonValue, _ := json.Marshal(auth)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusFound, w.Code)
}

func TestEditProfile(t *testing.T) {
	router := setupRouter()

	profile := profile{
		ID:   "1",
		Name: "Jane Doe",
		Age:  28,
		Bio:  "Product Manager",
		Pfp:  "profile2.jpg",
	}
	jsonValue, _ := json.Marshal(profile)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/profiles", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAddTag(t *testing.T) {
	router := setupRouter()

	tag := tag{
		Id:  "20",
		Tag: "golang",
	}
	jsonValue, _ := json.Marshal(tag)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/tags", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestQueryTag(t *testing.T) {
	router := setupRouter()

	id := id{
		Id: "1",
	}
	jsonValue, _ := json.Marshal(id)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/tags", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDeleteTag(t *testing.T) {
	router := setupRouter()

	tag := tag{
		Id:  "20",
		Tag: "golang",
	}
	jsonValue, _ := json.Marshal(tag)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/tags", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUpdateGeog(t *testing.T) {
	router := setupRouter()

	geog := geog{
		Id:   "1",
		Lat:  37.7749,
		Long: -122.4194,
	}
	jsonValue, _ := json.Marshal(geog)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/geog", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMetNumber(t *testing.T) {
	router := setupRouter()

	id := id{
		Id: "1",
	}
	jsonValue, _ := json.Marshal(id)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/metNumber", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMatches(t *testing.T) {
	router := setupRouter()

	id := id{
		Id: "1",
	}
	jsonValue, _ := json.Marshal(id)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/matches", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAddInterest(t *testing.T) {
	router := setupRouter()

	interest := idPair{
		IdFrom: "1",
		IdTo:   "2",
	}
	jsonValue, _ := json.Marshal(interest)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/interest", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAddNotInterest(t *testing.T) {
	router := setupRouter()

	interest := idPair{
		IdFrom: "1",
		IdTo:   "2",
	}
	jsonValue, _ := json.Marshal(interest)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/notInterest", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetMessage(t *testing.T) {
	router := setupRouter()

	msg := message{
		IdFrom:   "1",
		IdTo:     "2",
		TimeSent: time.Now(),
	}
	jsonValue, _ := json.Marshal(msg)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/message", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSendMessage(t *testing.T) {
	router := setupRouter()

	msg := message{
		IdFrom:   "1",
		IdTo:     "2",
		Msg:      "Hello!",
		TimeSent: time.Now(),
	}
	jsonValue, _ := json.Marshal(msg)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/message", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetChat(t *testing.T) {
	router := setupRouter()

	id := id{
		Id: "1",
	}
	jsonValue, _ := json.Marshal(id)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/chat", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
