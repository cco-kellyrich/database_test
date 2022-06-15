package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/jcelliott/lumber"
)

const Version = "1.0.0"

type (
	Logger interface {
		Fatal(string, ...interface{})
		Error(string, ...interface{})
		Warn(string, ...interface{})
		Info(string, ...interface{})
		Debug(string, ...interface{})
		Trace(string, ...interface{})
	}

	Driver struct {
		mutex   sync.Mutex
		mutexes map[string]*sync.Mutex
		dir     string
		log     Logger
	}
)
type Options struct {
	Logger
}

func New(dir string, options *Options) (*Driver, error) {
	dir = filepath.Clean(dir)

	opts := Options{}
	if options != nil {
		opts = *options
	}
	if opts.Logger == nil {
		opts.Logger = lumber.NewConsoleLogger((lumber.INFO))

	}
	driver := Driver{
		dir:     dir,
		mutexes: make(map[string]*sync.Mutex),
		log:     opts.Logger,
	}
	if _, err := os.Stat(dir); err == nil {
		opts.Logger.Debug("Using '%s'(database already exists)\n", dir)
		return &driver, nil
	}
	opts.Logger.Debug("creating the database at '%s'...\n", dir)
	return &driver, os.MkdirAll(dir, 0755)

}
func (d *Driver) Write(collection, resource string, v interface{}) error {
	if collection == "" {
		return fmt.Errorf("Missiong collection - no place to save the record!")
	}
	if resource == "" {
		return fmt.Errorf("Missing resource - unable to save record (no name)!")
	}
	mutex := d.getOrCreateMutex(collection)
	mutex.Lock()
	defer mutex.Unlock()

	dir := filepath.Join(d.dir, collection)
	fnlPath := filepath.Join(dir, resource+".json")
	tmpPath := fnlPath + ".tmp"

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	b, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return err
	}

	b = append(b, byte('\n'))
	if err := ioutil.WriteFile(tmpPath, b, 0644); err != nil {
		return err
	}
	return os.Rename(tmpPath, fnlPath)

}

func (d *Driver) Read(collection, resource string, v interface{}) error {
	if collection == "" {
		return fmt.Errorf("missing collection - no place to save record")
	}
	if resource == "" {
		return fmt.Errorf("missing resource - unable to save record (no name)!")
	}
	record := filepath.Join(d.dir, collection, resource)
	if _, err := stat(record); err != nil {
		return err
	}
	b, err := ioutil.ReadFile(record + ".json")
	if err != nil {
		return err
	}
	return json.Unmarshal(b, &v)
}

func (d *Driver) ReadAll(collection string) ([]string, error) {

	if collection == "" {
		return nil, fmt.Errorf("Missing collection - unable to read")

	}
	dir := filepath.Join(d.dir, collection)

	if _, err := stat(dir); err != nil {
		return nil, err
	}

	files, _ := ioutil.ReadDir(dir)

	var records []string

	for _, file := range files {
		b, err := ioutil.ReadFile(filepath.Join(dir, file.Name()))
		if err != nil {
			return nil, err
		}
		records = append(records, string(b))
	}
	return records, nil

}
func (d *Driver) Delete(collection, resource string) error {

	path := filepath.Join(collection, resource)
	mutex := d.getOrCreateMutex(collection)
	mutex.Lock()
	defer mutex.Unlock()

	dir := filepath.Join(d.dir, path)

	switch fi, err := stat(dir); {
	case fi == nil, err != nil:
		return fmt.Errorf("unable to find the directory named %v\n", path)

	case fi.Mode().IsDir():
		return os.RemoveAll(dir)

	case fi.Mode().IsRegular():
		return os.RemoveAll(dir + ".json")

	}
	return nil

}
func (d *Driver) getOrCreateMutex(collection string) *sync.Mutex {

	d.mutex.Lock()
	defer d.mutex.Unlock()
	m, ok := d.mutexes[collection]
	if !ok {
		m = &sync.Mutex{}
		d.mutexes[collection] = m

	}
	return m

}

func stat(path string) (fi os.FileInfo, err error) {
	if fi, err = os.Stat(path); os.IsNotExist(err) {
		fi, err = os.Stat(path + "json")
	}
	return
}

type Ingredients struct {
	Ingredient1 string
	Ingredient2 string
	Ingredient3 string
	Ingredient4 string
	Ingredient5 string
	Ingredient6 string
	Ingredient7 string
	Ingredient8 string
}

type Dinner struct {
	Name        string
	Type        string
	Ingredients Ingredients
	Rating      json.Number
}

func main() {
	dir := "./"

	db, err := New(dir, nil)
	if err != nil {
		fmt.Println("Error", err)
	}

	dinners := []Dinner{
		{"Spaghetti", "Italian", Ingredients{"Spaghetti noodles", "Sauce", "Parmesan Cheese", "", "", "", "", ""}, "8"},
		{"Stir Fry", "Asian", Ingredients{"Chicken", "Cauliflower", "Brocolli", "Mushrooms", "Asparagus", "Onion", "Rice", ""}, "7"},
		{"Taco", "Mexican", Ingredients{"Ground Beef", "Taco seasoning", "Taco shells", "Cheese", "Salsa", "", "", ""}, "9"},
		{"Pizza", "Italian", Ingredients{"Pizza Dough", "Pizza Sauce", "Mozzeralla Cheese", "Peperoni", "Sausage", "Black Olives", "", ""}, "10"},
	}

	for _, value := range dinners {
		db.Write("Dinner", value.Name, Dinner{
			Name:        value.Name,
			Type:        value.Type,
			Ingredients: value.Ingredients,
			Rating:      value.Rating,
		})
	}

	records, err := db.ReadAll("Dinner")
	if err != nil {
		fmt.Println("Error", err)
	}
	fmt.Println(records)

	alldinners := []Dinner{}

	for _, f := range records {
		dinnersFound := Dinner{}
		if err := json.Unmarshal([]byte(f), &dinnersFound); err != nil {
			fmt.Println("Error", err)
		}
		alldinners = append(alldinners, dinnersFound)

	}
	fmt.Println(alldinners)

	//if err := db.Delete( "Dinner", "Spaghetti"); err != nil{
	//	fmt.Println("error",err)
	//}
	//if err := db.Delete( "Dinner", ""); err != nil {
	//	fmt.Println("error", err)

	//}

}
