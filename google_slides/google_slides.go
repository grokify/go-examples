// Go example that covers:
// Quickstart: https://developers.google.com/slides/quickstart/go
// Basic writing: adding a text box to slide: https://developers.google.com/slides/samples/writing
// Using SDK: https://github.com/google/google-api-go-client/blob/master/slides/v1/slides-gen.go
// Creating and Managing Presentations https://developers.google.com/slides/how-tos/presentations
// Adding Shapes and Text to a Slide: https://developers.google.com/slides/how-tos/add-shape#example
package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path"
	"path/filepath"

	"github.com/grokify/gotilla/fmt/fmtutil"
	ou "github.com/grokify/oauth2util"
	oug "github.com/grokify/oauth2util/google"
	"github.com/joho/godotenv"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"google.golang.org/api/slides/v1"
)

// tokenCacheFile generates credential file path/filename.
// It returns the generated credential path/filename.
func tokenCacheFile() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	tokenCacheDir := filepath.Join(usr.HomeDir, ".credentials")
	os.MkdirAll(tokenCacheDir, 0700)
	return filepath.Join(tokenCacheDir,
		url.QueryEscape("slides.googleapis.com-go-quickstart.json")), err
}

func getConf() (*oauth2.Config, error) {
	return oug.ConfigFromEnv(oug.ClientSecretEnv,
		[]string{slides.DriveScope, slides.PresentationsScope})
}

func getSetClientWeb(tStore ou.TokenStoreFile) (*http.Client, error) {
	conf, err := getConf()
	if err != nil {
		return &http.Client{}, err
	}
	tok, err := ou.NewTokenFromWeb(conf)
	if err != nil {
		return &http.Client{}, err
	}
	tStore.Token = tok
	fmtutil.PrintJSON(tok)

	err = tStore.Write()
	if err != nil {
		return &http.Client{}, err
	}

	return conf.Client(context.Background(), tok), nil
}

func getTokenStore() ou.TokenStoreFile {
	credDir, err := ou.UserCredentialsDir()
	if err != nil {
		panic(err)
	}
	tokenPath := path.Join(credDir, "slides.googleapis.com-go-quickstart.json")
	return ou.NewTokenStoreFile(tokenPath)
}

func getClient(forceNewToken bool) (*http.Client, error) {
	err := godotenv.Load()
	if err != nil {
		return &http.Client{}, err
	}

	tStore := getTokenStore()
	err = tStore.Read()

	client := &http.Client{}

	if err != nil || forceNewToken {
		return getSetClientWeb(tStore)
	}

	conf, err := getConf()
	if err != nil {
		panic(err)
	}
	client = conf.Client(context.Background(), tStore.Token)
	return client, nil
}

func main() {
	forceNewToken := true

	client, err := getClient(forceNewToken)
	if err != nil {
		log.Fatal("Unable to get Client")
	}

	srv, err := slides.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve Slides Client %v", err)
	}

	psv := slides.NewPresentationsService(srv)

	pres := &slides.Presentation{Title: "GOLANG TEST PRES #2"}
	res, err := psv.Create(pres).Do()
	if err != nil {
		panic(err)
	}

	fmt.Printf("CREATED Presentation with Id %v\n", res.PresentationId)

	for i, slide := range res.Slides {
		fmt.Printf("- Slide #%d id %v contains %d elements.\n", (i + 1),
			slide.ObjectId,
			len(slide.PageElements))
	}

	pageId := res.Slides[0].ObjectId
	elementId := "MyTextBox_01"

	pt350 := &slides.Dimension{
		Magnitude: 350,
		Unit:      "PT"}

	requests := []*slides.Request{
		{
			CreateShape: &slides.CreateShapeRequest{
				ObjectId:  elementId,
				ShapeType: "TEXT_BOX",
				ElementProperties: &slides.PageElementProperties{
					PageObjectId: pageId,
					Size: &slides.Size{
						Height: pt350,
						Width:  pt350,
					},
					Transform: &slides.AffineTransform{
						ScaleX:     1.0,
						ScaleY:     1.0,
						TranslateX: 350.0,
						TranslateY: 100.0,
						Unit:       "PT",
					},
				},
			},
		},
		{
			InsertText: &slides.InsertTextRequest{
				ObjectId:       elementId,
				InsertionIndex: 0,
				Text:           "New Box Text Inserted!",
			},
		},
	}
	breq := &slides.BatchUpdatePresentationRequest{
		Requests: requests,
	}

	resu, err := psv.BatchUpdate(res.PresentationId, breq).Do()
	if err != nil {
		panic(err)
	}
	fmt.Println(resu.PresentationId)
}
