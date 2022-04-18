package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"sync"
)

var mutex_save sync.Mutex

func save_img(filename string, img []color.NRGBA) {
	mutex_save.Lock()

	if _, err := os.Stat(dist); os.IsNotExist(err) {
		os.Mkdir(dist, DIR_PERMISSION)
	}

	if _, err := os.Stat(dist + "/" + creation_name); os.IsNotExist(err) {
		os.Mkdir(dist+"/"+creation_name, DIR_PERMISSION)
	}

	file, err := os.Create(dist + "/" + creation_name + "/" + filename)
	if err != nil {
		fmt.Println("Couldn't open file for writing: ", err.Error())
		return
	}
	defer file.Close()

	painting := image.NewNRGBA(image.Rect(0, 0, width, height))
	for index, color := range img {
		painting.SetNRGBA(index%width, index/width, color)
	}

	err = png.Encode(file, painting)
	if err != nil {
		fmt.Println("Couldn't encode PNG: ", err.Error())
	}

	mutex_save.Unlock()
}

func save_config(filename string, config []byte) {
	if _, err := os.Stat(dist); os.IsNotExist(err) {
		os.Mkdir(dist, DIR_PERMISSION)
	}

	if _, err := os.Stat(dist + "/" + creation_name); os.IsNotExist(err) {
		os.Mkdir(dist+"/"+creation_name, DIR_PERMISSION)
	}

	file, err := os.Create(dist + "/" + creation_name + "/" + filename)
	if err != nil {
		fmt.Println("Couldn't open file for writing: ", err.Error())
		return
	}
	defer file.Close()

	_, err = file.Write(config)
	if err != nil {
		fmt.Println("Couldn't write config: ", err.Error())
	}
}

func save_resumable(filename string, config []byte) {
	if _, err := os.Stat(dist); os.IsNotExist(err) {
		os.Mkdir(dist, DIR_PERMISSION)
	}

	if _, err := os.Stat(dist + "/" + creation_name); os.IsNotExist(err) {
		os.Mkdir(dist+"/"+creation_name, DIR_PERMISSION)
	}

	file, err := os.Create(dist + "/" + creation_name + "/" + filename)
	if err != nil {
		fmt.Println("Couldn't open file for writing: ", err.Error())
		return
	}
	defer file.Close()

	_, err = file.Write(config)
	if err != nil {
		fmt.Println("Couldn't write config: ", err.Error())
	}
}
