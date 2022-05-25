package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/exp/maps"
)

type user struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}

type action struct {
	ID         int       `json:"id"`
	Type       string    `json:"type"`
	UserID     int       `json:"userId"`
	TargetUser int       `json:"targetUser"`
	CreatedAt  time.Time `json:"createdAt"`
}

var actions []action
var users []user

func main() {

	// Load data from .json files
	loadActions()
	loadUsers()

	// In case of non-existing users.json: comment above line and uncomment both below lines
	//extractUsersFromActions()
	//saveUsersToFile()

	// Define API's endpoints
	router := gin.Default()

	// Endpoint for Q1
	router.GET("/users/:id", getUserByID)

	// Endpoint for Q2
	router.GET("/users/actions/:id", getActionCountByUserID)

	// Endpoint for Q3
	router.GET("/actions/next/:type", getNextActionBreakdownByType)

	// Endpoint for Q4
	router.GET("/users/referralIndex", getUsersReferralIndex)

	// Start the server
	router.Run("localhost:8080")
}

func loadActions() {
	content, err := ioutil.ReadFile("./data/actions.json")
	if err != nil {
		log.Fatal("Error when opening actions file: ", err)
	}

	err = json.Unmarshal(content, &actions)
	if err != nil {
		log.Fatal("Error during Unmarshal() for actions: ", err)
	}
}

func loadUsers() {
	content, err := ioutil.ReadFile("./data/users.json")
	if err != nil {
		log.Fatal("Error when opening users file: ", err)
	}

	err = json.Unmarshal(content, &users)
	if err != nil {
		log.Fatal("Error during Unmarshal() for users: ", err)
	}
}

func extractUsersFromActions() {
	userMap := make(map[string]user)

	for _, action := range actions {
		userMap[strconv.Itoa(action.UserID)] = user{
			ID:        action.UserID,
			Name:      strings.Join([]string{"User", strconv.Itoa(action.UserID)}, ""),
			CreatedAt: action.CreatedAt,
		}
	}

	users = maps.Values(userMap)
}

func saveUsersToFile() {
	content, err := json.Marshal(users)
	if err != nil {
		log.Fatal("Error during Marshal() for users: ", err)
	}

	err = ioutil.WriteFile("./data/users.json", content, 0644)
	if err != nil {
		log.Fatal("Error when saving users file: ", err)
	}
}

func getUserByID(c *gin.Context) {
	receivedID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid id"})
		return
	}

	for _, user := range users {
		if user.ID == receivedID {
			c.IndentedJSON(http.StatusOK, user)
			return
		}
	}

	c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid id"})
}

func getActionCountByUserID(c *gin.Context) {
	receivedID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid id"})
		return
	}

	count := 0

	for _, action := range actions {
		if action.UserID == receivedID {
			count++
		}
	}

	c.IndentedJSON(http.StatusOK, gin.H{"count": count})
}

func getNextActionBreakdownByType(c *gin.Context) {
	receivedType := c.Param("type")

	// Map to keep track of users who have executed an action of the relevant type
	userMap := make(map[int]time.Time)

	// Map to count how many times different actions were taken next
	nextActionMap := make(map[string]float64)

	totalNextActions := 0

	for _, action := range actions {
		actionDate, ok := userMap[action.UserID]
		if ok {
			// Only take into account actions that are taken within a 24h window
			if action.CreatedAt.Sub(actionDate).Hours() <= 24.0 {
				nextActionMap[action.Type] = nextActionMap[action.Type] + 1
				totalNextActions++
			}
			delete(userMap, action.UserID)
		} else {
			if action.Type == receivedType {
				userMap[action.UserID] = action.CreatedAt
			}
		}
	}

	// Calculate probabilities
	for index, actionCount := range nextActionMap {
		nextActionMap[index] = actionCount / float64(totalNextActions)
	}

	c.IndentedJSON(http.StatusOK, nextActionMap)
}

func getUsersReferralIndex(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, gin.H{"message": "TBD"})
}
