package app

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path"
	"strings"
)

type TryAvatars []Avatar

func (a TryAvatars) GetAvatarURL(u ChatUser) (string, error) {
	for _, avatar := range a {
		if url, err := avatar.GetAvatarURL(u); err == nil {
			return url, err
		}
	}

	return "", ErrNoAvatarURL
}

var avatars TryAvatars = TryAvatars{
	UseFileSystemAvatar,
	UseAuthAvatar,
	UseGravatar,
}

// ErrNoAvatar is the error that is returned when the
// Avatar instance is unable to provide an avatar URL.
var ErrNoAvatarURL = errors.New("chat: Unable to get an avatar URL.")

// Avatar represents types capable of representing
// user profile pictures.
type Avatar interface {
	// GetAvatarURL gets the Avatar URL for the specified client,
	// or returns an error if something goes wrong.
	// ErrNoAvatarURL is returned if the object is unable to get
	// a URL for the speicified client.
	GetAvatarURL(ChatUser) (string, error)
}

type AuthAvatar struct{}

func (AuthAvatar) GetAvatarURL(user ChatUser) (string, error) {
	url := user.AvatarURL()
	if len(url) == 0 {
		return "", ErrNoAvatarURL
	}

	return url, nil
}

type GravatarAvatar struct{}

func (g GravatarAvatar) GetAvatarURL(user ChatUser) (string, error) {
	// double slash before www means, if current site is running on http,
	// use http://www.gravatar.com, if current site is using https,
	// use https://www.gravatar.com
	return "//www.gravatar.com/avatar/" + user.UniqueID(), nil
}

type FileSystemAvatar struct{}

func (f FileSystemAvatar) GetAvatarURL(user ChatUser) (string, error) {
	if userId := user.UniqueID(); userId != "" {
		ext, err := f.getFileExt(userId)
		if err != nil {
			return "", ErrNoAvatarURL
		}
		return "/avatars/" + userId + ext, nil
	}
	return "", ErrNoAvatarURL
}

func (f FileSystemAvatar) getFileExt(userId string) (string, error) {
	files, err := ioutil.ReadDir("/home/gaurav/Documents/programs/myprojects/chat/avatars/")
	if err != nil {
		return "", err
	}

	for _, file := range files {
		if strings.HasPrefix(file.Name(), userId) {
			return path.Ext(file.Name()), nil
		}
	}

	return "", fmt.Errorf("No file found for user %s", userId)
}

// Since we didn't have to create an instance of AuthAvatar, no memory was allocated.
// This helps us save space when we have multiple rooms as no memory will be used
var (
	UseAuthAvatar       AuthAvatar
	UseGravatar         GravatarAvatar
	UseFileSystemAvatar FileSystemAvatar
)
