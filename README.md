# Back-end Coding Challenge

A simple REST API has been implemented in Go using the Gin Web Framework. A set of endpoints have been created to serve solutions to each question in the challenge.

## How to run the application

As pre-requisite, an installation of Go is required for this application to run.

The application can be started by executing the following command on a terminal, within the project's root folder:

```bash
go run .
```

This will start the server and make it available at http://localhost:8080

## Endpoints

The following API endpoints are accessible, and offer the solution for each question:

- Q1: http://localhost:8080/users/:userId
- Q2: http://localhost:8080/users/actions/:userId
- Q3: http://localhost:8080/actions/next/:actionType
- Q4: http://localhost:8080/users/referralIndex

## Implementation details

Disclaimer: this was the first Go application that I've worked with.
All the code can be found in the `main.go` file, and the `actions.json` and `users.json` can be found in the `data` folder.

The actions list is always sorted in chronological order before the server is started (specially relevant for Q3 and Q4). This is done using the `sort` package (part of the standard library) and it has complexity of O(n log(n)).

Q1 and Q2 are pretty straightforward, they are both implemented using a simple for loop. Both are of linear complexity: O(#users) for Q1 and O(#actions) for Q2

For Q3 it loops once through all the actions in the action list and it uses maps to store intermediate and final data, since insert, access and deletion operations in a map have constant time complexity. The solution is roughly of linear complexity too: O(#actions) + O(#unique actions).

Q4 was the trickiest one to figure out its most efficient implementation. This is the task I spend the most time on (see time breakdown for all tasks below), mostly on the planning stage: once a viable solution was found, the implementation itself was pretty fast. The solution also uses a map to store the intermediate and final values, and a slice to define a tree-like structure of father-child relationships. It loops once through all actions in the action list and creates the tree structure, then navigates the tree starting from the leaves (there is as many nodes as users that have been referred).

Again, the solution is roughly of linear complexity: O(#actions) + O(#referredUsers).

**Note**, that this returns a list of users (and their referral index) only if they have made a `REFER_USER` action at least once. It does not include users whose referral index is 0. In order to include those users too, one would need to loop through all users, adding O(#users) complexity to the solution.

**IMPORTANT:** the current implementation for Q4 does **NOT** provide correct results with the given data. The reason is that it was assumed that users that had been referred did not exist within the platform (that is could not be able to take actions) BEFORE being referred. A fast look at the data showed that there are several cases where this rule is broken, resulting in the algorithm ignoring indirect referrals that happened before the target user was referred. 

For example, if we want to calculate the referral index for user 104, we have the following information:

2021-10-30 20:52:45.978 +0000 UTC, user 883 refers to user 529
2021-11-14 01:22:09.253 +0000 UTC, user 104 refers to user 883
2021-11-28 19:57:18.742 +0000 UTC, user 529 refers to user 658

The correct index for user 104 should be **3**: 1 direct referral to 883, and 2 indirect referrals, but this solution returns an index of only **2**, because user 883, the target of user 104's referral, already took a referral action BEFORE, which is not added to 104 final index. 

The alternative solution for non-sorted action list (see `actions-not-sorted` branch) does return the correct result, and runs in similar linear complexity.

Additionally, a small couple of functions have been created to generate a `user.json` file out of the `actions.json` existing file. The functions' code is included in the final version, but the call to the functions have been commented. This was due to the fact that the `users.json` file was unavailable to download. It seems to be fixed now, and the file has been downloaded and added to the repository.

### Time breakdown

I've spend a total of around 8h working on this challenge.

- 2h setting up the REST API and reading the data files
- 20 min implementing Q1
- 20 min implementing Q2
- 1h implementing Q3
- 3h implementing Q4
- 1h reviewing and documenting

### Improvements

Here is a list of potential improvements that have not been done due to time constraints and unfamiliarity with the Go language.

- Refactor some of the code to other files/packages to improve readability.
- Add unit and/or E2E tests
- Add additional validation to requests that expect a parameter and improve error handling
- Better adhere to Go coding conventions

# Changelog

- Actions list now sorted in the main function, before initiating the server.
- In Q3, for ALL actions of the specified type we check their next action (if there is any). Previously, actions of the specified type that also were a next action were ignored.
- In Q3, the 24h window requirement between actions has been removed.