package main

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

type Album struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name      string             `json:"name" bson:"name"`
	ImagePath string             `json:"ipath" bson:"ipath"`
}

//CreateAlbum with create an empty album in the database "ImageAlbumStorage"
func CreateAlbum(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var album Album
	if err := json.NewDecoder(r.Body).Decode(&album); err != nil {
		log.Fatal(err)
	}
	fmt.Println(album.Name)

	collections, _ := client.Database("ImageAlbumStorage").ListCollectionNames(context.TODO(), bson.D{})

	fmt.Println(collections)
	for _, name := range collections {
		if album.Name == name {
			fmt.Println("Album already exists")
			return
		}
	}
	client.Database("ImageAlbumStorage").CreateCollection(context.TODO(), album.Name)
	fmt.Println("Created Album")
}

//DeleteAlbum will delete the album of the given name.
//It will first check if the album exists in the "ImageAlbumStorage" database
//If the album exists then it will delete that album
func DeleteAlbum(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var album Album
	if err := json.NewDecoder(r.Body).Decode(&album); err != nil {
		log.Fatal(err)
	}

	fmt.Println(album.Name)

	collections, _ := client.Database("ImageAlbumStorage").ListCollectionNames(context.TODO(), bson.D{})
	fmt.Println(collections)

	for _, name := range collections {
		if album.Name == name {
			if err := client.Database("ImageAlbumStorage").Collection(name).Drop(context.TODO()); err != nil {
				log.Fatal(err)
			}
			fmt.Println("Deleted the album : ", name)
			return
		}
	}
	fmt.Println("Album does not exists")
}

//CreateImageInAlbum will store an image in the album.
func CreateImageInAlbum(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var album Album
	if err := json.NewDecoder(r.Body).Decode(&album); err != nil {
		log.Fatal(err)
	}
	fmt.Println(album.Name)
	// Open file on disk.
	f, _ := os.Open(album.ImagePath)

	// Read entire JPG into byte slice.
	reader := bufio.NewReader(f)
	content, _ := ioutil.ReadAll(reader)

	// Encode as base64.
	encodedImage := base64.StdEncoding.EncodeToString(content)

	fmt.Println("ENCODED: " + encodedImage)
	client.Database("ImageAlbumStorage").Collection(album.Name).InsertOne(context.TODO(), album)

}

//DeleteImageInAlbum will delete the image present in the album.
func DeleteImageInAlbum(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	id, _ := primitive.ObjectIDFromHex(params["id"])
	var album Album
	if err := json.NewDecoder(r.Body).Decode(&album); err != nil {
		log.Fatal(err)
	}

	res, err := client.Database("ImageAlbumStorage").Collection(album.Name).DeleteOne(context.TODO(), bson.D{{"_id", id}})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(res.DeletedCount)

}

//ListImagesInAlbum will list all the images present in the album
func ListImagesInAlbum(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var album Album
	if err := json.NewDecoder(r.Body).Decode(&album); err != nil {
		fmt.Println("error while decoding")
		log.Fatal(err)
	}

	cursor, err := client.Database("ImageAlbumStorage").Collection(album.Name).Find(context.TODO(), bson.D{})
	if err != nil {
		log.Fatal(err)
	}

	var results []bson.M
	if err = cursor.All(context.TODO(), &results); err != nil {
		log.Fatal(err)
	}
	for _, result := range results {
		fmt.Println(result)
	}
}

//GetImageInAlbum will give the queried image present in the album
func GetImageInAlbum(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var album Album
	if err := json.NewDecoder(r.Body).Decode(&album); err != nil {
		log.Fatal(err)
	}
	params := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(params["id"])

	if err != nil {
		log.Fatal(err)
	}

	var result bson.M

	client.Database("ImageAlbumStorage").Collection(album.Name).FindOne(context.TODO(), bson.D{{"_id", id}}).Decode(&result)
	fmt.Println(result)
}

func main() {
	fmt.Println("Starting Application")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, _ = mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal(err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/album", CreateAlbum).Methods("POST")
	router.HandleFunc("/album", DeleteAlbum).Methods("DELETE")
	router.HandleFunc("/album/image", CreateImageInAlbum).Methods("POST")
	router.HandleFunc("/album/image/{id}", DeleteImageInAlbum).Methods("DELETE")
	router.HandleFunc("/album/image", ListImagesInAlbum).Methods("GET")
	router.HandleFunc("/album/image/{id}", GetImageInAlbum).Methods("GET")

	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal(err)
	}
}
