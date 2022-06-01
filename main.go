package main

import (
	"container/list"
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

	// In case of non-existing users.json file: comment above line and uncomment both below lines
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

	// Map to store, in chronological order, the actions taken by a specific user, being userId the map's key
	userActionsMap := make(map[int]*list.List)

	// Map to keep track of all the actions of the relevant type taken by each user
	userActionTypeMap := make(map[int][]*list.Element)

	// Map to count how many times different actions were taken next
	nextActionMap := make(map[string]float64)

	totalNextActions := 0
	var insertedAction *list.Element

	// Organize actions chronologically by user and collect references to actions of the specified type
	for _, action := range actions {
		userList, ok := userActionsMap[action.UserID]
		if ok {
			insertedAction = insertActionInOrder(userList, action)
		} else {
			userActionsMap[action.UserID] = list.New()
			insertedAction = userActionsMap[action.UserID].PushFront(action)
		}

		if action.Type == receivedType {
			userActionTypeMap[action.UserID] = append(userActionTypeMap[action.UserID], insertedAction)
		}
	}

	// Find "next" actions
	for _, receivedTypeActions := range userActionTypeMap {
		for _, actionReference := range receivedTypeActions {
			nextElement := actionReference.Next()
			if nextElement != nil {
				nextAction := nextElement.Value.(action)
				nextActionMap[nextAction.Type] = nextActionMap[nextAction.Type] + 1
				totalNextActions++
			}
		}
	}

	// Calculate probabilities
	for index, actionCount := range nextActionMap {
		nextActionMap[index] = actionCount / float64(totalNextActions)
	}

	c.JSON(http.StatusOK, nextActionMap)
}

func insertActionInOrder(l *list.List, a action) *list.Element {
	for e := l.Front(); e != nil; e = e.Next() {
		if e.Value.(action).CreatedAt.After(a.CreatedAt) {
			return l.InsertBefore(a, e)
		}
	}
	return l.PushBack(a)
}

func getUsersReferralIndex(c *gin.Context) {
	type userReferralNode struct {
		userID    int
		referer   int // ID of user who referred this user, -1 otherwise
		referrals int
	}

	// Map to store the referrals tree
	userReferralsTreeMap := make(map[int]userReferralNode)

	// Map to store the final referral index
	userReferralIndexMap := make(map[int]int)

	// Map to keep track of referred users with no direct referrals
	// In the tree structure, these would be the leaves
	treeLeavesMap := make(map[int]int)

	// Slice to keep track of the nodes to be visited when navigating the tree in reverse
	var referralsOrder []int

	// Create the referrals tree so each node has the amount of direct referrals for each user
	for _, action := range actions {
		if action.Type == "REFER_USER" && action.UserID != action.TargetUser {
			// Create or update node for referer user
			refererNode, ok := userReferralsTreeMap[action.UserID]
			if ok {
				refererNode.referrals++
				userReferralsTreeMap[action.UserID] = refererNode
			} else {
				userReferralsTreeMap[action.UserID] = userReferralNode{action.UserID, -1, 1}
			}

			// Referer user cannot be a leave, remove it from the treeLeavesMap
			delete(treeLeavesMap, action.UserID)

			// Create or update target user's info
			targetNode, ok := userReferralsTreeMap[action.TargetUser]
			if ok {
				// Update target user's referer
				targetNode.referer = action.UserID
				userReferralsTreeMap[action.TargetUser] = targetNode
			} else {
				// Create node for target user
				userReferralsTreeMap[action.TargetUser] = userReferralNode{action.TargetUser, action.UserID, 0}

				// Store target user as potential tree leave
				treeLeavesMap[action.TargetUser] = action.TargetUser
			}
		}
	}

	// Calculate the final referral index by adding indirect referrals to the already calculated direct ones

	// 1. Add tree leaves to referralsOrder slice as first nodes to visit
	for _, leaf := range treeLeavesMap {
		referralsOrder = append(referralsOrder, leaf)
	}

	// 2. Traverse the tree in reverse (starting from the leaves). In our case, we need to traverse the
	// referralsOrder slice in order, as leaves will be in the first positions, and parent nodes
	// will be appended at the end of the slice
	for i := 0; i < len(referralsOrder); i++ {
		userID := referralsOrder[i]

		nodeInfo := userReferralsTreeMap[userID]

		// If the node has a parent
		if nodeInfo.referer >= 0 {
			// Append it to the slice to be visited later
			referralsOrder = append(referralsOrder, nodeInfo.referer)

			// And increase their referrals amount with the current node's ones
			parentNode := userReferralsTreeMap[nodeInfo.referer]
			parentNode.referrals = parentNode.referrals + nodeInfo.referrals
			userReferralsTreeMap[nodeInfo.referer] = parentNode
		}

		// Update the referral index for the current node
		userReferralIndexMap[userID] = nodeInfo.referrals
	}

	c.JSON(http.StatusOK, userReferralIndexMap)
}
