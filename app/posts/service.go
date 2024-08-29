package posts

import (
	"errors"

	"gorm.io/gorm"
)

type PostService struct {
	DB *gorm.DB
}

func (s *PostService) CreatePost(post *Post) error {
	return s.DB.Create(post).Error
}

func (s *PostService) GetPostByID(id uint) (*Post, error) {
	var post Post
	if err := s.DB.First(&post, id).Error; err != nil {
		return nil, err
	}
	return &post, nil
}

func (s *PostService) GetAllPosts() ([]Post, error) {
	var posts []Post
	if err := s.DB.Find(&posts).Error; err != nil {
		return nil, err
	}
	return posts, nil
}

func (s *PostService) UpdatePost(post *Post) error {
	if err := s.DB.Save(post).Error; err != nil {
		return err
	}
	return nil
}

func (s *PostService) DeletePost(id uint) error {
	result := s.DB.Delete(&Post{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("post not found")
	}
	return nil
}
