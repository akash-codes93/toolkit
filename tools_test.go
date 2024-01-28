package toolkit

import (
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
)

func TestTools_RandomString(t *testing.T) {
	var testTools Tools

	s := testTools.RandomString(10)
	if len(s) != 10 {
		t.Error("length is not right")
	}
}

var uploadTests = []struct {
	name          string
	allowedTypes  []string
	renameFile    bool
	errorExpected bool
}{
	{name: "allowed no rename", allowedTypes: []string{"image/jpeg", "image/png"}, renameFile: false, errorExpected: false},
	{name: "allowed rename", allowedTypes: []string{"image/jpeg", "image/png"}, renameFile: true, errorExpected: false},
	{name: "not allowed", allowedTypes: []string{"image/png"}, renameFile: true, errorExpected: true},
}

func TestTools_UploadFiles(t *testing.T) {
	for _, e := range uploadTests {
		// setup a pipe to avoid buffering
		pr, pw := io.Pipe()
		writer := multipart.NewWriter(pw)

		wg := sync.WaitGroup{}
		wg.Add(1)

		go func() {
			defer writer.Close()
			defer wg.Done()

			// create the form data filed "file"
			part, err := writer.CreateFormFile("file", "./testdata/img.jpeg")
			if err != nil {
				t.Error(err)
			}

			f, err := os.Open("./testdata/img.jpeg")
			if err != nil {
				t.Error(err)
			}
			defer f.Close()
			img, _, err := image.Decode(f)
			if err != nil {
				t.Error(err)
			}
			quality := jpeg.Options{
				Quality: 50,
			}
			err = jpeg.Encode(part, img, &quality)
			if err != nil {
				t.Error(err)
			}
		}()
		// read from the pipe which recieves data
		request := httptest.NewRequest("POST", "/", pr)
		request.Header.Add("Content-Type", writer.FormDataContentType())

		var testTools Tools
		testTools.AllowedFileTypes = e.allowedTypes
		uploadedFiles, err := testTools.UploadedFiles(request, "./testdata/uploads/", e.renameFile)
		if err != nil && !e.errorExpected {
			t.Error(err)
		}

		if !e.errorExpected {
			if _, err := os.Stat(fmt.Sprintf("./testdata/uploads/%s", uploadedFiles[0].NewFileName)); os.IsNotExist(err) {
				t.Errorf("%s expected the file to exists: %s", e.name, err.Error())
			}
			_ = os.Remove(fmt.Sprintf("./testdata/uploads/%s", uploadedFiles[0].NewFileName))
		}

		if !e.errorExpected && err != nil {
			t.Errorf("%s expected but non recieved", e.name)

		}
		wg.Wait()

	}
}

func TestTools_CreateDirIfNotExists(t *testing.T) {
	var testTools Tools

	err := testTools.CreateDirIfNotExists("./testdata/myDir")
	if err != nil {
		t.Error(err)
	}

	err = testTools.CreateDirIfNotExists("./testdata/myDir")
	if err != nil {
		t.Error(err)
	}

	_ = os.Remove("./testdata/myDir")

}
