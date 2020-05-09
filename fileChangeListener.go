package main

import (
	"fmt"
	"os"
	"path/filepath"
	"os/exec"
	"gopkg.in/fsnotify.v1"
	"strings"
)

var watcher *fsnotify.Watcher // global watchers

var files []string // global slice to store files

// main
func main() {

	folderToWatch := "./" // current folder as default

	// if argument was passed, watch the folder
	if len(os.Args) > 1 {
		folderToWatch = os.Args[1]
	}

	// if target folders exists
	file, err := os.Stat(folderToWatch)
	if err != nil {
		fmt.Println("Provided folder doesn't exist ", file)
		return
	}

	fmt.Println("watching: ", folderToWatch)

	// creates a new file watcher
	watcher, _ = fsnotify.NewWatcher()
	defer watcher.Close()

	// starting at the root of the project, walk each file/directory searching for
	// directories to watch
	if err := filepath.Walk(folderToWatch, watchDir); err != nil {
		fmt.Println("ERROR", err)
	}

	_files, err := GetFilesRecursively(folderToWatch) // get all files from the path
	if err != nil {
		return
	}

	addFiles(_files); // add files to the list

	//
	done := make(chan bool)

	//
	go func() {
		for {
			select {
			// watch for events
			case event := <-watcher.Events:
				// fmt.Printf("EVENT! %#v\n", event)
				addNewWatcher(event)
				notify(event)

				// watch for errors
			case err := <-watcher.Errors:
				fmt.Println("ERROR", err)
			}
		}
	}()

	<-done
}

/**
 * get files recursively from the folder and its subfolder
 * var root 	path
 */
func GetFilesRecursively(root string) ([]string, error) {
    var _files []string

    // walk all directories but collect only files
    err := filepath.Walk(root, func(path string, f os.FileInfo, err error) error {

		// add to list if its a file
		if !f.IsDir() {
			_files = append(_files, path)
		}
        return nil
    })

    return _files, err
}


// watchDir gets run as a walk func, searching for directories to add watchers to
func watchDir(path string, fi os.FileInfo, err error) error {
	// since fsnotify can watch all the files in a directory, watchers only need
	// to be added to each nested directory
	if fi.Mode().IsDir() {
		return watcher.Add(path)
	}

	return nil
}

/**
 * add files to the global list
 * and call the action
 * var newFiles 	array of paths
 */
func addFiles(newFiles []string) {
	for _, file := range newFiles {
		// only if not hidden folder/file
		if !strings.Contains(file, "/.") {
			files = append(files, file) // add file to the list
			execScript(file, "add"); //	call exec function for the current file
		}
	}
	// printFiles()
}

/**
 * remvoe file from the list
 * and call script
 *
 */
func removefiles(path string) {
	for index, file := range files {
		if strings.HasPrefix(file, path) {
			fmt.Println("removed", file)
			files = remove(files, index)
			execScript(file, "remove"); // call exeternal script
		}
	}
	// printFiles()
}

/**
 * remove string item from slice and keep the order
 *
 */
func remove(slice []string, s int) []string {
    return append(slice[:s], slice[s+1:]...)
}

/**
 *	log all files stored in global var
 */
func printFiles() {
	fmt.Println("files :")
	for index, file := range files {
		fmt.Println(index, ": ", file)
	}
	fmt.Println()
}

/**
 *	Add new watcher
 */
func addNewWatcher(event fsnotify.Event) {
	// if file/folder was created
	if event.Op == 1 {

		//  verify if file exists
		file, err := os.Stat(event.Name)
		if err != nil {
			return
		}

		mode := file.Mode()

		// if its a directory, get all files from it, and create a new watcher
		if mode.IsDir() {
			// add a watcher to new subfolder
			if err := filepath.Walk(event.Name, watchDir); err != nil {
				fmt.Println("ERROR", err)
			} else {
				fmt.Println("ADDED new watcher to folder ", event.Name)

				_newFiles, _ := GetFilesRecursively(event.Name) // get all files within the directory
				addFiles(_newFiles) // add files to global variable
			}
		} else {

			// append files to global files
			addFiles([]string{event.Name}) // add files to global variable
		}
	}
	return
}

/**
 * full path of the file
 * action - add / remove
 * script to call wp loop media file [add|remove] /path/to/file
 */
func execScript(path string, action string) {
	out, err := exec.Command("bash", "-c", " wp loop media file " + action + " " + path).Output()
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	fmt.Printf("output: %s\n", out)

}

/**
 *	notification from watcher, handles only delete,
 * 	as the create is handled within adding watcher function
 */
func notify(event fsnotify.Event) {
	// err => file doesnt exist => deleted
	_, err := os.Stat(event.Name)
	if err != nil {
		removefiles(event.Name) // remove file from the global array
	}

	return
}
