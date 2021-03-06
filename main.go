package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
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

	// In case of non-existing users.json file: comment above line and uncomment both below lines
	//extractUsersFromActions()
	//saveUsersToFile()

	// Sort actions list chronologically
	sort.Slice(actions, func(i, j int) bool { return actions[i].CreatedAt.Before(actions[j].CreatedAt) })

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

	// Experimental function, not to be used in prod
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id"})
		return
	}

	for _, user := range users {
		if user.ID == receivedID {
			c.JSON(http.StatusOK, user)
			return
		}
	}

	c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id"})
}

func getActionCountByUserID(c *gin.Context) {
	receivedID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id"})
		return
	}

	count := 0

	for _, action := range actions {
		if action.UserID == receivedID {
			count++
		}
	}

	c.JSON(http.StatusOK, gin.H{"count": count})
}

func getNextActionBreakdownByType(c *gin.Context) {
	receivedType := c.Param("type")

	// Map to keep track of users who have executed an action of the relevant type
	userMap := make(map[int]time.Time)

	// Map to count how many times different actions were taken next
	nextActionMap := make(map[string]float64)

	totalNextActions := 0

	for _, action := range actions {
		_, ok := userMap[action.UserID]
		if ok {
			nextActionMap[action.Type] = nextActionMap[action.Type] + 1
			totalNextActions++
			delete(userMap, action.UserID)
		}

		if action.Type == receivedType {
			userMap[action.UserID] = action.CreatedAt
		}
	}

	// Calculate probabilities
	for index, actionCount := range nextActionMap {
		nextActionMap[index] = actionCount / float64(totalNextActions)
	}

	c.JSON(http.StatusOK, nextActionMap)
}

func getUsersReferralIndex(c *gin.Context) {
	type userReferralInfo struct {
		userID  int
		referer int // ID of user who referred this user
	}

	// Map to keep track of user referrals
	userReferralIndexMap := make(map[int]int)

	// Slice to keep track of the order in which users are referred, and their referrers
	var referralsOrder []userReferralInfo

	// Populate the map with each user's direct referrals
	for _, action := range actions {
		if action.Type == "REFER_USER" {
			// Increase the amount of users referred by action.UserID
			userReferralIndexMap[action.UserID] = userReferralIndexMap[action.UserID] + 1

			// Keep track of the order in which users are referred and
			// define action.UserID as the referrer for action.TargetUser
			referralsOrder = append(referralsOrder, userReferralInfo{action.TargetUser, action.UserID})
		}
	}

	// Add indirect referrals to calculate final referral index
	// Traverse the referralsOrder slice in reverse, skipping the last position
	// as it would have referrals = 0
	for i := len(referralsOrder) - 2; i >= 0; i-- {
		targetInfo := referralsOrder[i]

		// Update referer's referral index to include referred's referrals (indirect referrals)
		userReferralIndexMap[targetInfo.referer] = userReferralIndexMap[targetInfo.referer] + userReferralIndexMap[targetInfo.userID]
	}

	c.JSON(http.StatusOK, userReferralIndexMap)
}
