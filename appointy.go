/*
 * appointy instagram back-end REST end point API server
 * @version : V1.0.0 (SEMVER)
 * @author : Rohan Reddy Melachervu
 * @profession : CSE Undergrad
 * @institution : VIT university, Chennai Campus
 * @vit-registration-number : 19BCE1191
 * @email : rohanreddy.melachervu2019@vitstudent.ac.in
 * @created : 09-10-2021
 *
 * Copyright (C) 2021  Rohan Reddy Melachervu
 *
 *  This program is free software: you can redistribute it and/or modify
 *  it under the terms of the GNU General Public License as published by
 *  the Free Software Foundation, either version 3 of the License, or
 *  (at your option) any later version.

 *  This program is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *  GNU General Public License for more details.

 *  You should have received a copy of the GNU General Public License
 *  along with this program.  If not, see <https://www.gnu.org/licenses/>.
 *
 */
package main

import (
	"context"
	"fmt"
	"log"
	"path"
	"strconv"
	"strings"
	"time"

	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"math/rand"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// var mongoOptions *options.ClientOptions
var MongoClient *mongo.Client
var Users *mongo.Collection
var Posts *mongo.Collection

type User struct {
	Id       string `bson:"text"`
	Name     string `bson:"text"`
	Email    string `bson:"text"`
	Password string `bson:"text"`
}

type Post struct {
	Id       int64
	Caption  string `bson:"text"`
	ImageURL string `bson:"text"`
	PostedTS string `bson:"text"`
	UserId   string `bson:"text"`
}

var ctx context.Context

func main() {
	//connectToMongoDB()
	rand.Seed(time.Now().UnixNano())
	getUsersCollection()
	getPostsCollection()

	// REST end point router
	http.HandleFunc("/users", createUser)
	http.HandleFunc("/posts/users", getPostsByUserID)
	http.HandleFunc("/posts/users/", getAPagefullPostsByUserID)
	http.HandleFunc("/posts", createPost)
	http.HandleFunc("/users/", getUserByUserID)
	http.HandleFunc("/posts/", getPostByPostID)
	http.ListenAndServe(":8080", nil)
}

func init() {
	MongoClient = connectToMongoDB()
	listDBs()
}

func dryRunCode() {
	user := User{
		Id:       "venkat_reddy",
		Name:     "Venkateswar Reddy M",
		Email:    "venkat@brillium.tech",
		Password: "venkat@123",
	}

	var _ = user
	InsertUser(user)
	postId := time.Now().UnixNano() / (1 << 22)
	ts := time.Now().Format(time.RFC850)
	post := Post{
		Id:       postId,
		Caption:  "First Appointy Post",
		ImageURL: "https://www.linkedin.com/in/vmelachervu",
		PostedTS: ts,
	}
	post.PostedTS = time.Now().Format(time.RFC3339)
	InsertPost(post, user.Id)
	foundUser, err := findUsersByUserID("venkat_reddy")
	var _ = err
	fmt.Println(foundUser)

	foundPost, err := findPostByPostID(4037200794235010051)
	fmt.Printf("Found a post: %+v\n", foundPost)

	foundPosts, err := ListAllUserPosts(user)
	var _ = foundPosts
}

/*
 * Connects to mongo
 *
 */
func connectToMongoDB() *mongo.Client {
	MongoClient, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	err = MongoClient.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	// defer MongoClient.Disconnect(ctx)
	// check the mongo is really connected
	err = MongoClient.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to MongoDB!")
	return MongoClient
}

func listDBs() {
	databases, err := MongoClient.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Println(databases)
	var _ = databases
}

/*
 * reads existing users coolection from mongo
 *
 */
func getUsersCollection() {

	Users = MongoClient.Database("appointyMongoDB").Collection("users")
}

/*
 * reads existing posts coolection from mongo
 *
 */
func getPostsCollection() {

	Posts = MongoClient.Database("appointyMongoDB").Collection("posts")
}

/*
 * Creates user in the mongo
 * POST <base_url>/users with User JSON body
 *
 */
func InsertUser(user User) (*mongo.InsertOneResult, error) {
	insertResult, err := Users.InsertOne(ctx, bson.D{
		{Key: "Id", Value: user.Id},
		{Key: "Name", Value: user.Name},
		{Key: "Email", Value: user.Email},
		{Key: "Password", Value: user.Password},
	})
	if err != nil {
		fmt.Println(err)
	}
	return insertResult, err
}

/*
 * Finds user by UserID in the mongo
 * GET <base_url>/users/{UserId}
 *
 */
func findUsersByUserID(userID string) ([]bson.M, error) {
	if userID == "" {
		return []bson.M{}, nil
	}
	filterCursor, err := Users.Find(ctx, bson.M{"Id": userID})
	var foundUsers []bson.M
	if err = filterCursor.All(ctx, &foundUsers); err != nil {
		log.Fatal(err)
	}
	// fmt.Println(foundUsers)
	if err != nil {
		log.Fatal(err)
	}
	return foundUsers, err
}

/*
 * Finds post by PostId in the mongo
 * GET <base_url>/posts/{PostID}
 *
 */
func findPostByPostID(postID int) ([]bson.M, error) {
	filterCursor, err := Posts.Find(ctx, bson.M{"Id": postID})
	var foundPosts []bson.M
	if err = filterCursor.All(ctx, &foundPosts); err != nil {
		log.Fatal(err)
	}
	return foundPosts, err
}

/*
 * Gets all posts of a user in the mongo
 * GET <base_url>/posts/users/{UserID}
 *
 */
func ListAllUserPosts(user User) ([]bson.M, error) {
	filterCursor, err := Posts.Find(ctx, bson.M{"UserId": user.Id})
	var foundPosts []bson.M
	if err = filterCursor.All(ctx, &foundPosts); err != nil {
		log.Fatal(err)
	}
	// fmt.Println(foundPosts)
	if err != nil {
		log.Fatal(err)
	}
	return foundPosts, err
}

/*
 * Gets a pageful of posts of a user in the mongo
 * GET <base_url>/posts/users/{UserID}/{pageNum}
 *
 */
func ListAPagefulUserPosts(user User, pageNum int64) ([]bson.M, error) {
	opts := options.Find()
	var pageSize int64 = 3
	opts.SetLimit(pageSize)
	var skip int64 = pageSize
	if pageNum > 1 {
		skip = skip * (pageNum - 1)
	} else {
		skip = 0
	}
	opts.SetSkip(skip)
	opts.SetSort(bson.D{{"PostedTS", -1}})
	currentPostsPage, err := Posts.Find(ctx, bson.M{"UserId": user.Id}, opts)
	var sortedCurrentPage []bson.M
	if err = currentPostsPage.All(ctx, &sortedCurrentPage); err != nil {
		log.Fatal(err)
	}
	if err != nil {
		log.Fatal(err)
	}
	return sortedCurrentPage, err
}

/*
 * Creates a post for a user in the mongo
 * POST <base_url>/posts
 *
 */
func InsertPost(post Post, userID string) (*mongo.InsertOneResult, error) {

	insertResult, err := Posts.InsertOne(ctx, bson.D{
		{Key: "Id", Value: time.Now().UnixNano() / (1 << 22)},
		{Key: "Caption", Value: post.Caption},
		{Key: "ImageURL", Value: post.ImageURL},
		{Key: "PostedTS", Value: time.Now().Format(time.RFC850)},
		{Key: "UserId", Value: userID},
	})
	if err != nil {
		fmt.Println(err)
	}
	return insertResult, err
}

/*
 * Creates a user in the mongo
 * POST <base_url>/users
 *
 */
func createUser(rw http.ResponseWriter, request *http.Request) {
	decoder := json.NewDecoder(request.Body)

	var t User
	err := decoder.Decode(&t)

	if err != nil {
		fmt.Fprintf(rw, "Unable to create User - reason: %s\n", err)
		return
	}
	existingUser, err := findUsersByUserID(t.Id)
	if err != nil {
		fmt.Fprintf(rw, "Unable to create User - reason: %s\n", err)
		return
	}
	if len(existingUser) > 0 {
		fmt.Fprintf(rw, "User \"%s\" already exists\n", t.Id)
		return
	}
	Sha2512hasher := sha512.New()
	Sha2512hasher.Write([]byte(t.Password))
	passHash := hex.EncodeToString(Sha2512hasher.Sum(nil))
	t.Password = passHash
	insertedUser, err := InsertUser(t)
	if err != nil {
		fmt.Fprintf(rw, "Unable to create User - reason: %s\n", err)
		return
	}
	fmt.Fprintf(rw, "UserId %s created successfully for %s \n", t.Id, t.Name)
	var _ = insertedUser
}

/*
 * Gets the first pageful of posts for a given user in the mongo
 * POST <base_url>/posts/users/{UserId}
 *
 */
func getPostsByUserID(rw http.ResponseWriter, request *http.Request) {
	userID := path.Base(request.RequestURI)
	user := User{Id: userID}
	userPosts, err := ListAPagefulUserPosts(user, 1)
	if len(userPosts) == 0 {
		fmt.Fprint(rw, "User has no posts\n")
		return
	}
	var _ = err
	posts, _ := json.Marshal(userPosts)
	// fmt.Println(string(posts))
	fmt.Fprintf(rw, "Users posts are: \n")
	fmt.Fprintf(rw, "%s\n", string(posts))
}

/*
 * Gets a requested pageful of posts for a given user in the mongo
 * POST <base_url>/posts/users/{UserId}/{PageID}
 *
 */
func getAPagefullPostsByUserID(rw http.ResponseWriter, request *http.Request) {
	pageID := path.Base(request.RequestURI)
	page, err := strconv.ParseInt(pageID, 10, 64)
	var _ = pageID
	urlPart := strings.Split(request.RequestURI, "/")
	var _ = urlPart
	userID := urlPart[3]
	user := User{Id: userID}
	userPosts, err := ListAPagefulUserPosts(user, page)
	if len(userPosts) == 0 {
		fmt.Fprint(rw, "User has no posts\n")
		return
	}
	var _ = err
	posts, _ := json.Marshal(userPosts)
	// fmt.Println(string(posts))
	fmt.Fprintf(rw, "Users posts are: \n")
	fmt.Fprintf(rw, "%s\n", string(posts))
}

/*
 * Creates a post for a given user in the mongo - expects the userID in the JSON along with post details
 * POST <base_url>/posts
 *
 */
func createPost(rw http.ResponseWriter, request *http.Request) {
	decoder := json.NewDecoder(request.Body)

	var t Post
	err := decoder.Decode(&t)

	if err != nil {
		fmt.Fprintf(rw, "Unable to post - reason: %s\n", err)
		return
	}
	insertedPost, err := InsertPost(t, t.UserId)
	if err != nil {
		fmt.Fprintf(rw, "Unable to post - reason: %s\n", err)
		return
	}
	fmt.Fprintf(rw, "Post \"%s\" created successfully for %s \n", t.Caption, t.UserId)
	var _ = insertedPost
}

/*
 * Gets user details from mongo for a given userID
 * GET <base_url>/users/{userID}
 *
 */
func getUserByUserID(rw http.ResponseWriter, request *http.Request) {
	userID := path.Base(request.RequestURI)
	user := User{Id: userID}
	userReturned, err := findUsersByUserID(user.Id)
	if len(userReturned) == 0 {
		fmt.Fprintf(rw, "The user \"%s\" doesn't exist", userID)
		return
	}
	var _ = err
	userz, _ := json.Marshal(userReturned)
	// fmt.Println(string(posts))
	fmt.Fprintf(rw, "User found: \n")
	fmt.Fprintf(rw, "%s\n", string(userz))
}

/*
 * Gets a post details from mongo for a given POSTID
 * GET <base_url>/posts/{PostID}
 *
 */
func getPostByPostID(rw http.ResponseWriter, request *http.Request) {
	postID := path.Base(request.RequestURI)
	id, err := strconv.Atoi(postID)
	if err != nil {
		fmt.Fprintf(rw, "The PostId %d does not exist. Please retry with valid PostId", id)
		return
	}
	foundPost, err := findPostByPostID(id)
	if len(foundPost) == 0 {
		fmt.Fprintf(rw, "The Post for \"%d\" doesn't exist", id)
		return
	}
	var _ = err
	post, _ := json.Marshal(foundPost)
	fmt.Fprintf(rw, "Post message is: \n")
	fmt.Fprintf(rw, "%s\n", string(post))
}
