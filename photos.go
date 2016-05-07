package myapp

import (
	"crypto/sha1"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"google.golang.org/cloud/storage"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine"
)

// I added a parameter.. The request
func uploadPhoto(src multipart.File, hdr *multipart.FileHeader, c *http.Cookie, req http.Request) *http.Cookie {
	defer src.Close()
	//Limit kinds of files you can upload
	if(hdr.Filename != ".jpg" ||
	   hdr.Filename != ".png" ||
	   hdr.Filename != ".txt"){
		return c
	}

	m := Model(c)

	var fName string

	// get filename with extension and store it in fName.
	//just setting up a basic file structure
	//bucket/ userName/encryptedFilename
	fName =  m.Name + getSha(src) + filepath.Ext(hdr.Filename)


	//grabbing context for error checking
	ctx := appengine.NewContext(req)

	//creating a new client from our context
	client, err := storage.NewClient(ctx)
	if err != nil{
		log.Errorf(ctx, "Error in main client err")
	}
	defer client.Close()

	//grabbing a client fronm our specific bucket
	bucket := client.Bucket(gcsBucket)

	//making a new gcs writer
	writer := bucket.Object(fName).NewWriter(ctx)

	//making the file public
	writer.ACL = []storage.ACLRule{
		{storage.AllUsers, storage.RoleReader},
	}

	//setting the type of the file png/jpg/txt
	writer.ContentType = hdr.Filename

	//writing the file to the gcs bucket
	//NOT SURE IF I'M ALLOWED TO CONVERT OUR FILE TO []byte
	_, err = writer.Write([]byte(src))

	if(err != nil){
		log.Errorf(ctx, "uploadPhoto: unable to write data to bucket")
		return
	}

	err = writer.Close()
	if(err != nil){
		log.Errorf(ctx, "uploadPhoto closing writer")
		return
	}

	return addPhoto(fName, filepath.Ext(hdr.Filename), c)
}

//Stores the file path inside the Model
func addPhoto(fName string, ext string, c *http.Cookie) *http.Cookie {
	//Get Model for m.Pictures
	m := Model(c)
	//If the file is an image with jpg or png extension, put it in m.Pictures
	if ext == ".jpg" || ext == ".png"{
		m.Pictures = append(m.Pictures, fName)
	}
	//Store the file path in string slice, m.Files
	m.Files = append(m.Files, fName)
	//Get id from old Model and update the cookie with updated model
	xs := strings.Split(c.Value, "|")
	id := xs[0]
	cookie := currentVisitor(m, id)
	return cookie
}

//encryption stuff
func getSha(src multipart.File) string{
	h := sha1.New()
	io.Copy(h, src)
	return fmt.Sprintf("%x", h.Sum(nil))
}