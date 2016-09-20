/*
Copyright 2016 The Fission Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"io/ioutil"
	"os"
	"path"

	log "github.com/Sirupsen/logrus"
)

type requestType int

const (
	READ requestType = iota
	WRITE
	DELETE
)

type (
	fileStore struct {
		root           string // abs path of root of filestore
		requestChannel chan fileStoreRequest
	}

	fileStoreRequest struct {
		requestType
		fileName        string // relative path
		fileContents    []byte
		responseChannel chan fileStoreResponse
	}

	fileStoreResponse struct {
		error
		fileContents []byte
	}
)

func makeFileStore(path string) *fileStore {
	fileStore := &fileStore{
		root:           path,
		requestChannel: make(chan fileStoreRequest),
	}
	go fileStore.fileStoreService()
	return fileStore
}

func (fs *fileStore) fileStoreService() {
	for {
		req := <-fs.requestChannel
		response := &fileStoreResponse{}

		log.WithFields(log.Fields{"file": req.fileName, "type": req.requestType}).Debug("fileStore request")
		switch req.requestType {
		case READ:
			response.fileContents, response.error = ioutil.ReadFile(path.Join(fs.root, req.fileName))
		case WRITE:
			response.error = ioutil.WriteFile(path.Join(fs.root, req.fileName), req.fileContents, 0600)
		case DELETE:
			response.error = os.Remove(path.Join(fs.root, req.fileName))
		default:
			log.Panic("bad request")
		}
		req.responseChannel <- *response
	}
}

func (fs *fileStore) read(fileName string) ([]byte, error) {
	req := fileStoreRequest{
		requestType:     READ,
		fileName:        fileName,
		responseChannel: make(chan fileStoreResponse),
	}
	fs.requestChannel <- req
	response := <-req.responseChannel
	return response.fileContents, response.error
}

func (fs *fileStore) write(fileName string, contents []byte) error {
	req := fileStoreRequest{
		requestType:     WRITE,
		fileName:        fileName,
		fileContents:    contents,
		responseChannel: make(chan fileStoreResponse),
	}
	fs.requestChannel <- req
	response := <-req.responseChannel
	return response.error
}

func (fs *fileStore) delete(fileName string) error {
	req := fileStoreRequest{
		requestType:     DELETE,
		fileName:        fileName,
		responseChannel: make(chan fileStoreResponse),
	}
	fs.requestChannel <- req
	response := <-req.responseChannel
	return response.error
}
