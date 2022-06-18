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

var (
	errNoOperationFlag = errors.New("-operation flag has to be specified")
	errNoFileName      = errors.New("-fileName flag has to be specified")
	errNoItem          = errors.New("-item flag has to be specified")
	errIncorrectItem   = errors.New("incorrect format of an item")
	errNoId            = errors.New("-id flag has to be specified")
)

type User struct {
	Id    string `json:"id"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

type Arguments map[string]string

func Perform(args Arguments, writer io.Writer) error {

	var users []User
	itemPos := -1
	foundItem := false

	//валидация параметров
	if args["fileName"] == "" {
		return errNoFileName
	}
	if args["operation"] == "" {
		return errNoOperationFlag
	}
	if args["operation"] != "add" && args["operation"] != "list" && args["operation"] != "findById" && args["operation"] != "remove" {
		return fmt.Errorf("Operation %v not allowed!", args["operation"])
	}

	f, err := os.OpenFile(args["fileName"], os.O_RDWR|os.O_CREATE, 0660)
	if err != nil {
		return err
	}
	defer f.Close()

	buff, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	//если файл не пустой,то скидываем данные в слайс
	if len(buff) != 0 {
		err = json.Unmarshal(buff, &users)
		if err != nil {
			return err
		}
	}

	switch args["operation"] {
	case "list":
		list, err := json.Marshal(users)
		if err != nil {
			return err
		}
		writer.Write(list)
	case "add":
		if args["item"] == "" {
			return errNoItem
		}

		if !json.Valid([]byte(args["item"])) {
			return errIncorrectItem
		}
		var item User
		json.Unmarshal([]byte(args["item"]), &item)

		for i, v := range users {
			if v.Id == item.Id {
				itemPos = i
				foundItem = true
			}
		}

		if !foundItem {
			var itemtoAdd User
			json.Unmarshal([]byte(args["item"]), &itemtoAdd)
			users = append(users, itemtoAdd)
			_, err = f.Seek(0, io.SeekStart)
			if err != nil {
				return err
			}
			err = f.Truncate(0)
			if err != nil {
				return err
			}

			stringToWright, err := json.Marshal(users)
			if err != nil {
				return err
			}
			f.WriteString(string(stringToWright))

		} else {
			writer.Write([]byte(fmt.Sprintf("Item with id %s already exists", item.Id)))

		}

	case "remove":
		foundItem = false
		if args["id"] == "" {
			return errNoId
		}
		for i, v := range users {
			if v.Id == args["id"] {
				foundItem = true
				itemPos = i
			}
		}
		if foundItem {
			users[itemPos] = users[len(users)-1]
			users = users[:len(users)-1]
			if err != nil {
				return err
			}
			_, err = f.Seek(0, io.SeekStart)
			if err != nil {
				return err
			}
			err = f.Truncate(0)
			if err != nil {
				return err
			}
			stringToWright, err := json.Marshal(users)
			if err != nil {
				return err
			}
			f.WriteString(string(stringToWright))

		} else {
			writer.Write([]byte(fmt.Sprintf("Item with id %s not found", args["id"])))
		}
	case "findById":
		var stringToWright []byte
		if args["id"] == "" {
			return errNoId
		}

		for i, v := range users {
			if v.Id == args["id"] {

				stringToWright, err = json.Marshal(users[i])
				if err != nil {
					return err
				}
			}
		}

		if len(stringToWright) != 0 {
			writer.Write(stringToWright)
		} else {
			writer.Write([]byte(""))
		}
	}
	return nil
}

func parseArgs() Arguments {

	var op, fn, id, item string
	flag.StringVar(&op, "operation", "", "operation to do")
	flag.StringVar(&fn, "fileName", "", "file to deal with")
	flag.StringVar(&item, "item", "", "a json item")
	flag.StringVar(&id, "id", "", "id of an item in json file")
	flag.Parse()
	args := Arguments{"id": id, "item": item, "operation": op, "fileName": fn}

	return args
}

func main() {
	err := Perform(parseArgs(), os.Stdout)
	if err != nil {
		panic(err)
	}
}
