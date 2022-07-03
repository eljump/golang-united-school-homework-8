package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

type Arguments map[string]string

type User struct {
	Id    string `json:"id"`
	Email string `json:"email"`
	Age   uint   `json:"age"`
}

var specifiedErrors = map[string]string{
	"fileName":  "-fileName flag has to be specified",
	"operation": "-operation flag has to be specified",
	"id":        "-id flag has to be specified",
	"item":      "-item flag has to be specified",
}

func Perform(args Arguments, writer io.Writer) error {
	var specErr error

	specErr = checkErrors(args, "operation")
	if specErr != nil {
		return specErr
	}

	specErr = checkErrors(args, "fileName")
	if specErr != nil {
		return specErr
	}

	file, specErr := os.OpenFile(args["fileName"], os.O_RDWR|os.O_CREATE, 0755)

	if specErr != nil {
		return specErr
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}(file)

	switch args["operation"] {
	case "add":
		specErr = checkErrors(args, "item")
		if specErr != nil {
			return specErr
		}
		addError := add(args["item"], file, writer)
		if addError != nil {
			writer.Write([]byte(addError.Error()))
		}
	case "list":
		list(file, writer)
	case "findById":
		specErr = checkErrors(args, "id")
		if specErr != nil {
			return specErr
		}
		user, findError := find(args["id"], file)
		if findError != nil {
			return findError
		}

		if user == nil {
			writer.Write([]byte(""))
		} else {
			value, _ := json.Marshal(user)
			writer.Write(value)
		}
	case "remove":
		specErr = checkErrors(args, "id")
		if specErr != nil {
			return specErr
		}
		user, removeErr := find(args["id"], file)
		if removeErr != nil {
			return removeErr
		}
		if user == nil {
			writer.Write([]byte(fmt.Errorf("Item with id %s not found", args["id"]).Error()))
		} else {
			remove(user, file)
		}
	default:
		return fmt.Errorf("Operation %s not allowed!", args["operation"])
	}

	return nil
}

func list(file *os.File, writer io.Writer) {
	content, _ := ioutil.ReadAll(file)
	writer.Write(content)
	file.Seek(0, 0)
}

func remove(user *User, file *os.File) error {
	defer file.Seek(0, 0)
	content, _ := ioutil.ReadAll(file)

	var users []User
	_ = json.Unmarshal(content, &users)
	for key, iterationUser := range users {
		if iterationUser.Id == user.Id {
			users = append(users[:key], users[key+1:]...)
		}
	}

	list, _ := json.Marshal(users)
	file.Truncate(0)
	file.Seek(0, 0)
	file.Write(list)

	return nil
}

func add(item string, file *os.File, writer io.Writer) error {
	user := &User{}
	err := json.Unmarshal([]byte(item), user)
	if err != nil {
		return err
	}

	currentUser, err := find(user.Id, file)
	if currentUser != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("Item with id %s already exists", user.Id)
	}

	fileContent, _ := ioutil.ReadAll(file)
	var users []User
	_ = json.Unmarshal(fileContent, &users)

	users = append(users, *user)
	byteUsers, _ := json.Marshal(users)

	file.Write(byteUsers)

	list(file, writer)

	return nil
}

func find(id string, file *os.File) (*User, error) {
	defer file.Seek(0, 0)
	fileContent, _ := ioutil.ReadAll(file)

	var users []User
	err := json.Unmarshal(fileContent, &users)
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		if user.Id == id {
			return &user, nil
		}
	}

	return nil, nil
}

func checkErrors(args Arguments, name string) error {
	if _, ok := args[name]; !ok || args[name] == "" {
		return errors.New(specifiedErrors[name])
	}
	return nil
}

func parseArgs() Arguments {

	id := flag.String("id", "", "user id")
	operation := flag.String("operation", "", "operation type")
	item := flag.String("item", "", "user")
	fileName := flag.String("fileName", "", "fileName")

	flag.Parse()

	return Arguments{
		"id":        *id,
		"operation": *operation,
		"item":      *item,
		"fileName":  *fileName,
	}
}

func main() {
	err := Perform(parseArgs(), os.Stdout)
	if err != nil {
		panic(err)
	}
}
