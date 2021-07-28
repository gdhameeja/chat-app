package app

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path"
	"strings"
)

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
	GetAvatarURL(c *client) (string, error)
}

type AuthAvatar struct{}

func (AuthAvatar) GetAvatarURL(c *client) (string, error) {
	// avatar_url is set in auth while setting the cookie
	if url, ok := c.userData["avatar_url"]; ok {
		if urlStr, ok := url.(string); ok {
			return urlStr, nil
		}
	}

	return "", ErrNoAvatarURL
}

type GravatarAvatar struct{}

func (g GravatarAvatar) GetAvatarURL(c *client) (string, error) {
	if userId, ok := c.userData["userId"]; ok {
		if userIdStr, ok := userId.(string); ok {
			// double slash before www means, if current site is running on http,
			// use http://www.gravatar.com, if current site is using https,
			// use https://www.gravatar.com
			return "//www.gravatar.com/avatar/" + userIdStr, nil
		}
	}

	return "", ErrNoAvatarURL
}

type FileSystemAvatar struct{}

func (f FileSystemAvatar) GetAvatarURL(c *client) (string, error) {
	if userId, ok := c.userData["userId"]; ok {
		if userIdStr, ok := userId.(string); ok {
			ext, err := f.getFileExt(userIdStr)
			if err != nil {
				return "", ErrNoAvatarURL
			}
			return "/avatars/" + userIdStr + ext, nil
		}
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
