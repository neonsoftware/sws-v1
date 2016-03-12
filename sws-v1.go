package sws

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type AllTemplateData struct {
	BoardData   BoardTemplateData    `json:"board_data"`
	ModulesData []ModuleTemplateData `json:"modules_data"`
}

type ModuleTemplateData struct {
	Inclusions          string `json:"inclusions"`
	Declarations        string `json:"declarations"`
	Initializations     string `json:"initializations"`
	Sending             string `json:"sending"`
	SendValueDocHeader  string `json:"send_value_doc_header"`
	SendValueDocMessage string `json:"send_value_doc_message"`
}

type BoardTemplateData struct {
	Inclusions      string `json:"inclusions"`
	Declarations    string `json:"declarations"`
	Initializations string `json:"initializations"`
	Receiving       string `json:"receiving"`
	PreSending      string `json:"presending"`
	PostSending     string `json:"postsending"`
}

type templates struct {
	Mix  string `json:"mix"`
	DotH string `json:".h"`
}

type osc_raw_info struct {
	FilesToCopy []string  `json:"files_to_copy"`
	Templates   templates `json:"templates"`
}

type images struct {
	Hundred string `json:"100"`
	Img     string `json:"img"`
	Uniform string `json:"uniform"`
}

type Manifest struct {
	Hws        string       `json:"hws"`
	Sws        string       `json:"sws"`
	Type       string       `json:"type"`
	Class      string       `json:"class"`
	Sixe_x     int          `json:"size_x"`
	Sixe_y     int          `json:"size_y"`
	ResMin     int          `json:"res_min"`
	OscRawInfo osc_raw_info `json:"osc_raw"`
	Images     images       `json:"images"`
}

type Module struct {
	Element_id  int    `json:"element_id"`
	Abs_id      int    `json:"element_id"`
	Abs_coord_x int    `json:"abs_coord_x"`
	Abs_coord_y int    `json:"abs_coord_y"`
	Rel_id      int    `json:"rel_id"`
	Rel_coord_x int    `json:"rel_coord_x"`
	Rel_coord_y int    `json:"rel_coord_y"`
	Type        string `json:"type"`
	Size_x      int    `json:"size_x"`
	Size_y      int    `json:"size_y"`
	Name        string `json:"name"`
}

type board struct {
	Id             int      `json:"id"`
	Type           string   `json:"type"`
	Coord_x        int      `json:"coord_x"`
	Coord_y        int      `json:"coord_y"`
	ProductId      string   `json:"product_id"`
	Elems_per_side int      `json:"elems_per_side"`
	Modules        []Module `json:"elements"`
}

type arch struct {
	Boards_num_x int `json:"boards_x"`
	Boards_num_y int `json:"boards_y"`
}

type map_config struct {
	Hws    string  `json:"hws"`
	Sws    string  `json:"sws"`
	Arch   arch    `json:"arch"`
	Boards []board `json:"boards"`
}

type Device struct {
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Uuid        string     `json:"uuid"`
	MapConfig   map_config `json:"map"`
}

func UnmarshalAllManifests(folder_to_scan string) ([]Manifest, error) {
	manifests := []Manifest{}
	pattern := filepath.Join(folder_to_scan, "*/sw/manifest.json")
	fmt.Println("Searching global pattern : ", pattern)
	manifest_files, _ := filepath.Glob(pattern)
	for _, file_path := range manifest_files {
		new_manifest, err := UnmarshalManifest(file_path)
		if err != nil {
			fmt.Println("error unmarshaling file ", file_path)
			return []Manifest{}, err
		}
		manifests = append(manifests, *new_manifest)
	}
	return manifests, nil
}

func UnmarshalManifestsByClass(folder_to_scan string, class_name string) ([]Manifest, error) {

	selected_manifests := []Manifest{}
	manifests, err := UnmarshalAllManifests(folder_to_scan)
	if err != nil {
		return []Manifest{}, err
	}
	for _, man := range manifests {
		if man.Class == class_name {
			selected_manifests = append(selected_manifests, man)
		}
	}
	return selected_manifests, nil
}

// TODO unused
func UnmarshalManifestByType(folder_to_scan string, module_type string) (*Manifest, error) {

	file_path := filepath.Join(folder_to_scan, module_type, "sw/manifest.json")
	return UnmarshalManifest(file_path)
}

func UnmarshalManifest(file_path string) (*Manifest, error) {

	manifestObj := &Manifest{}

	// Reading the manifest file
	file, err := ioutil.ReadFile(file_path)
	if err != nil {
		fmt.Println("Error while reading file " + file_path)
		fmt.Print(err)
		return manifestObj, nil
	}

	// Unmarshaling device json
	err = json.Unmarshal(file, &manifestObj)
	if err != nil {
		fmt.Println("Error unmarshaling" + file_path)
		fmt.Print(err)
		return manifestObj, nil
	}

	return manifestObj, nil
}

func UnmarshalDevice(file_path string) (*Device, error) {

	deviceObj := &Device{}

	// Reading the manifest file
	file, err := ioutil.ReadFile(file_path)
	if err != nil {
		fmt.Println("Error while reading file " + file_path)
		fmt.Print(err)
		return deviceObj, nil
	}

	// Unmarshaling the outer device json, map will remain a string
	err = json.Unmarshal(file, &deviceObj)
	if err != nil {
		fmt.Println("Error unmarshaling" + file_path)
		fmt.Print(err)
		return deviceObj, nil
	}

	return deviceObj, nil
}

func Unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	fmt.Println("unzipping " + src)
	if err != nil {
		fmt.Println("0")
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			fmt.Println("1")
			return err
		}
		defer rc.Close()

		path := filepath.Join(dest, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			f, err := os.OpenFile(
				path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				fmt.Println("2")
				return err
			}
			defer f.Close()

			_, err = io.Copy(f, rc)
			if err != nil {
				fmt.Println("3")
				return err
			}
		}
	}

	return nil
}

func Zipit(source, target string) error {
	zipfile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	info, err := os.Stat(source)
	if err != nil {
		return nil
	}

	var baseDir string
	if info.IsDir() {
		baseDir = filepath.Base(source)
	}

	filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		if baseDir != "" {
			header.Name = filepath.Join(baseDir, strings.TrimPrefix(path, source))
		}

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(writer, file)
		return err
	})

	return err
}

// TODO rename me and iterate on the struct
func GetBoardTemplList() []string {
	//new_list :=
	return []string{"Inclusions", "Declarations", "Initializations", "Receiving", "PreSending", "PostSending"}
}
