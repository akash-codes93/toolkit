package toolkit

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const randomStringSource = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_+"

type Tools struct {
	MaxFileSize        int
	AllowedFileTypes   []string
	MAXJSONSize        int
	AllowUnknownFields bool
}

// generates a random string og length n, using randomSourceString
func (t *Tools) RandomString(n int) string {
	s, r := make([]rune, n), []rune(randomStringSource)

	for i := range s {
		p, _ := rand.Prime(rand.Reader, len(r))
		x, y := p.Uint64(), uint64(len(r))
		s[i] = r[x%y]
	}

	return string(s)
}

// uploaded file is a struct used to save information.
type UploadedFile struct {
	NewFileName      string
	OriginalFileName string
	FileSize         int64
}

func (t *Tools) UploadedFiles(r *http.Request, uploadDir string, rename ...bool) ([]*UploadedFile, error) {
	renameFile := true
	if len(rename) > 0 {
		renameFile = rename[0]
	}

	var uploadedFiles []*UploadedFile

	if t.MaxFileSize == 0 {
		t.MaxFileSize = 1024 * 1024 * 1024
	}

	err := r.ParseMultipartForm(int64(t.MaxFileSize))

	if err != nil {
		return nil, errors.New("the uploaded file is too big")
	}
	// fmt.Println(r.MultipartForm.File) // dict of slices
	for _, fHeaders := range r.MultipartForm.File {
		for _, hdr := range fHeaders {
			err = func() error {
				var uploadedFile UploadedFile
				infile, err := hdr.Open()

				if err != nil {
					return err
				}
				defer infile.Close()

				buff := make([]byte, 512)
				_, err = infile.Read(buff)
				if err != nil {
					return err
				}

				allowed := false
				fileType := http.DetectContentType(buff)

				if len(t.AllowedFileTypes) > 0 {
					for _, x := range t.AllowedFileTypes {
						// ignore case while comparison
						if strings.EqualFold(fileType, x) {
							allowed = true
						}
					}
				} else {
					allowed = true
				}

				if !allowed {
					return errors.New("the file type is not permitted")
				}
				_, err = infile.Seek(0, 0)
				if err != nil {
					return err
				}
				// renaming the file
				if renameFile {
					uploadedFile.NewFileName = fmt.Sprintf("%s%s", t.RandomString(25), filepath.Ext(hdr.Filename))
				} else {
					uploadedFile.NewFileName = hdr.Filename
				}

				uploadedFile.OriginalFileName = hdr.Filename

				var outfile *os.File
				defer outfile.Close()

				if outfile, err = os.Create(filepath.Join(uploadDir, uploadedFile.NewFileName)); err != nil {
					return err
				} else {
					fileSize, err := io.Copy(outfile, infile)
					if err != nil {
						return err
					}
					uploadedFile.FileSize = fileSize
				}

				uploadedFiles = append(uploadedFiles, &uploadedFile)

				return nil

			}()
			if err != nil {
				return uploadedFiles, err
			}
		}
	}

	return uploadedFiles, nil
}

// create a dir if not exists and create all the necessary parents
func (t *Tools) CreateDirIfNotExists(path string) error {
	const mode = 0755
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.MkdirAll(path, mode)
		if err != nil {
			return err
		}
	}
	return nil
}

// slugify function enables you to generate a slug from a string
func (t *Tools) Slugify(s string) (string, error) {
	if s == "" {
		return "", errors.New("empty string not permitted")
	}

	var re = regexp.MustCompile(`[^a-z\d]+`)
	slug := strings.Trim(re.ReplaceAllString(strings.ToLower(s), "-"), "-")
	if len(slug) == 0 {
		return "", errors.New("after removing chars slug is 0 length")
	}

	return slug, nil
}

// JSON representation of response to be send
// type JSONResponse struct {
// 	Error   bool        `json:"error"`
// 	Message string      `json:"message:"`
// 	Data    interface{} `json:"data,omitempty"`
// }

// func (t *Tools) ReadJSON(w http.ResponseWriter, r *http.Request, data interface{}) error {
// 	maxBytes := 1024 * 1024 // one meg
// 	if t.MAXJSONSize != 0 {
// 		maxBytes = t.MAXJSONSize
// 	}

// 	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

// 	dec := json.NewDecoder(r.Body)

// 	if
// }

// gives the list of files that have changed in a folder
func (t *Tools) FindChangesInFolder(path string) ([]string, error) {

	diff := []string{}
	gitFile := path + ".json"
	files, _ := getAllFiles(path)
	root := createMerkelTree(files)

	if !checkFileExists(gitFile) {
		fmt.Println("Folder was not git checked, adding " + gitFile + "...")
		bytes, err := json.Marshal(root)
		if err != nil {
			return diff, err
		}
		// fmt.Println(string(bytes))
		_ = os.WriteFile(gitFile, bytes, 0644)
		return diff, nil
	}
	// reading bytes from a file
	bytes, err := os.ReadFile(gitFile)
	if err != nil {
		return diff, err
	}
	// unmarshall to tree
	var oldRoot MerkelNode
	err = json.Unmarshal(bytes, &oldRoot)
	if err != nil {
		return diff, err
	}

	diff = checkDifferentFiles(&oldRoot, root)
	if len(diff) > 1 && diff[len(diff)-1] == diff[len(diff)-2] {
		diff = diff[:len(diff)-1]
	}

	return diff, nil
}
