package handlers

import "am-keramika-backend/storage"

var imageStorage storage.ImageStorage

func SetImageStorage(store storage.ImageStorage) {
	imageStorage = store
}

func getImageStorage() storage.ImageStorage {
	return imageStorage
}
