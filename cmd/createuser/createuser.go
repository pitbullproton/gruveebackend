package createuser

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/firestore"
	"github.com/pixelogicdev/gruveebackend/pkg/firebase"
	"github.com/pixelogicdev/gruveebackend/pkg/social"
)

var firestoreClient *firestore.Client

func init() {
	log.Println("CreateUser Initialized")
}

// CreateUser will write a new Firebase user to Firestore
func CreateUser(writer http.ResponseWriter, request *http.Request) {
	// Initialize
	initWithEnvErr := initWithEnv()
	if initWithEnvErr != nil {
		http.Error(writer, initWithEnvErr.Error(), http.StatusInternalServerError)
		log.Printf("CreateUser [initWithEnv]: %v", initWithEnvErr)
		return
	}

	var createUserReq social.CreateUserReq

	jsonDecodeErr := json.NewDecoder(request.Body).Decode(&createUserReq)
	if jsonDecodeErr != nil {
		http.Error(writer, jsonDecodeErr.Error(), http.StatusInternalServerError)
		log.Printf("CreateUser [social.CreateUserReq Decoder]: %v", jsonDecodeErr)
		return
	}

	// Get Document references for social platform
	socialPlatDocRef := firestoreClient.Doc(createUserReq.SocialPlatformPath)
	if socialPlatDocRef == nil {
		http.Error(writer, jsonDecodeErr.Error(), http.StatusInternalServerError)
		log.Printf("CreateUser [social.CreateUserReq Decoder]: %v", jsonDecodeErr)
		return
	}

	// Create Firestore user
	firestoreUser := firebase.FirestoreUser{
		Email:                   createUserReq.Email,
		ID:                      createUserReq.ID,
		Playlists:               []*firestore.DocumentRef{},
		PreferredSocialPlatform: socialPlatDocRef,
		ProfileImage:            createUserReq.ProfileImage,
		SocialPlatforms:         []*firestore.DocumentRef{socialPlatDocRef},
		DisplayName:             createUserReq.DisplayName,
		Username:                createUserReq.Username,
	}

	// Write FirestoreUser to Firestore
	_, writeErr := firestoreClient.Collection("users").Doc(firestoreUser.ID).Set(context.Background(), firestoreUser)
	if writeErr != nil {
		http.Error(writer, writeErr.Error(), http.StatusInternalServerError)
		log.Printf("CreateUser [fireStore Set]: %v", writeErr)
		return
	}

	// Return Firestore User
	writer.WriteHeader(http.StatusOK)
	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(firestoreUser)
}

// Helpers

// initWithEnv takes our yaml env variables and maps them properly.
// Unfortunately, we had to do this is main because in init we weren't able to access env variables
func initWithEnv() error {
	// Get paths
	var currentProject string

	if os.Getenv("ENVIRONMENT") == "DEV" {
		currentProject = os.Getenv("FIREBASE_PROJECTID_DEV")
	} else if os.Getenv("ENVIRONMENT") == "PROD" {
		currentProject = os.Getenv("FIREBASE_PROJECTID_PROD")
	}

	// Initialize Firestore
	client, err := firestore.NewClient(context.Background(), currentProject)
	if err != nil {
		return fmt.Errorf("CreateUser [Init Firestore]: %v", err)
	}

	firestoreClient = client
	return nil
}
