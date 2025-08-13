package blog_posts_test

import (
	"testing"
	"time"

	"base/app/models"
	"base/test"

	"github.com/stretchr/testify/assert"
)

func TestBlogPostModels(t *testing.T) {
	helper := test.SetupTest(t)
	defer helper.TeardownTest()

	// Auto-migrate BlogPost table for testing
	err := helper.DB.AutoMigrate(&models.BlogPost{})
	assert.NoError(t, err)

	t.Run("BlogPost model operations comprehensive coverage", func(t *testing.T) {
		t.Run("create and basic operations", func(t *testing.T) {
			// Test BlogPost creation
			blogpost := &models.BlogPost{
				Title: nil,
				Content: "Test Content",
				Price: nil,
				AuthorId: ,
				PublishedAt: true,
				IsFeatured: true,
				Count: nil,
				
			}
			err := helper.DB.Create(blogpost).Error
			assert.NoError(t, err)
			assert.NotZero(t, blogpost.Id)
			assert.NotZero(t, blogpost.CreatedAt)
			assert.NotZero(t, blogpost.UpdatedAt)

			// Test find by ID
			var foundBlogPost models.BlogPost
			err = helper.DB.First(&foundBlogPost, blogpost.Id).Error
			assert.NoError(t, err)
			assert.Equal(t, blogpost.Id, foundBlogPost.Id)
			assert.Equal(t, nil, foundBlogPost.Title)
			assert.Equal(t, "Test Content", foundBlogPost.Content)
			assert.Equal(t, nil, foundBlogPost.Price)
			assert.Equal(t, , foundBlogPost.AuthorId)
			assert.Equal(t, true, foundBlogPost.PublishedAt)
			assert.Equal(t, true, foundBlogPost.IsFeatured)
			assert.Equal(t, nil, foundBlogPost.Count)
			
		})

		t.Run("BlogPost model methods", func(t *testing.T) {
			blogpost := &models.BlogPost{
				Title: nil,
				Content: "Test Content",
				Price: nil,
				AuthorId: ,
				PublishedAt: true,
				IsFeatured: true,
				Count: nil,
				
			}
			err := helper.DB.Create(blogpost).Error
			assert.NoError(t, err)

			// Test TableName
			assert.Equal(t, "blog_posts", blogpost.TableName())

			// Test GetId
			assert.Equal(t, blogpost.Id, blogpost.GetId())

			// Test GetModelName
			assert.Equal(t, "blogpost", blogpost.GetModelName())

			// Test ToListResponse
			listResponse := blogpost.ToListResponse()
			assert.NotNil(t, listResponse)
			assert.Equal(t, blogpost.Id, listResponse.Id)
			assert.Equal(t, nil, listResponse.Title)
			assert.Equal(t, "Test Content", listResponse.Content)
			assert.Equal(t, nil, listResponse.Price)
			assert.Equal(t, , listResponse.AuthorId)
			assert.Equal(t, true, listResponse.PublishedAt)
			assert.Equal(t, true, listResponse.IsFeatured)
			assert.Equal(t, nil, listResponse.Count)
			

			// Test ToResponse
			response := blogpost.ToResponse()
			assert.NotNil(t, response)
			assert.Equal(t, blogpost.Id, response.Id)
			assert.Equal(t, nil, response.Title)
			assert.Equal(t, "Test Content", response.Content)
			assert.Equal(t, nil, response.Price)
			assert.Equal(t, , response.AuthorId)
			assert.Equal(t, true, response.PublishedAt)
			assert.Equal(t, true, response.IsFeatured)
			assert.Equal(t, nil, response.Count)
			

			// Test Preload
			query := blogpost.Preload(helper.DB)
			assert.NotNil(t, query)
		})

		t.Run("BlogPost soft delete", func(t *testing.T) {
			blogpost := &models.BlogPost{
				Title: nil,
				Content: "Test Content",
				Price: nil,
				AuthorId: ,
				PublishedAt: true,
				IsFeatured: true,
				Count: nil,
				
			}
			err := helper.DB.Create(blogpost).Error
			assert.NoError(t, err)

			// Test timestamps
			assert.True(t, blogpost.CreatedAt.Before(time.Now().Add(time.Second)))
			assert.True(t, blogpost.UpdatedAt.Before(time.Now().Add(time.Second)))

			// Test soft delete
			err = helper.DB.Delete(blogpost).Error
			assert.NoError(t, err)

			// Should not find deleted record
			var foundBlogPost models.BlogPost
			err = helper.DB.First(&foundBlogPost, blogpost.Id).Error
			assert.Error(t, err)

			// Should find with Unscoped
			err = helper.DB.Unscoped().First(&foundBlogPost, blogpost.Id).Error
			assert.NoError(t, err)
			assert.Equal(t, blogpost.Id, foundBlogPost.Id)
			assert.NotZero(t, foundBlogPost.DeletedAt)
		})

		t.Run("validation of required fields", func(t *testing.T) {
			blogpost := &models.BlogPost{}
			err := helper.DB.Create(blogpost).Error
			