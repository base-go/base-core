package blog_posts_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"base/app/blog_posts"
	"base/app/models"
	"base/core/emitter"
	"base/test"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestBlogPostControllers(t *testing.T) {
	helper := test.SetupTest(t)
	defer helper.TeardownTest()

	// Auto-migrate BlogPost table for testing
	err := helper.DB.AutoMigrate(&models.BlogPost{})
	assert.NoError(t, err)

	// Create service and controller
	service := blog_posts.NewBlogPostService(
		helper.DB,
		emitter.New(),
		nil, // storage can be nil for basic tests
		helper.Logger,
	)
	controller := blog_posts.NewBlogPostController(service, nil)

	// Setup router
 
	router := router.New()
	api := router.Group("/api")
	controller.Routes(api)

	t.Run("BlogPost controller operations comprehensive coverage", func(t *testing.T) {
		t.Run("Create BlogPost", func(t *testing.T) {
			createReq := models.CreateBlogPostRequest{
				Title: nil,
				Content: "Test Content",
				Price: nil,
				AuthorId: ,
				PublishedAt: true,
				IsFeatured: true,
				Count: nil,
				
			}

			jsonData, _ := json.Marshal(createReq)
			req, _ := http.NewRequest("POST", "/api/blog-posts", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusCreated, w.Code)

			var response models.BlogPostResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.NotZero(t, response.Id)
			assert.Equal(t, nil, response.Title)
			assert.Equal(t, "Test Content", response.Content)
			assert.Equal(t, nil, response.Price)
			assert.Equal(t, , response.AuthorId)
			assert.Equal(t, true, response.PublishedAt)
			assert.Equal(t, true, response.IsFeatured)
			assert.Equal(t, nil, response.Count)
			
		})

		t.Run("Get BlogPost by ID", func(t *testing.T) {
			// Create test blogpost
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

			req, _ := http.NewRequest("GET", fmt.Sprintf("/api/blog-posts/%d", blogpost.Id), nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response models.BlogPostResponse
			err = json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, blogpost.Id, response.Id)
			assert.Equal(t, nil, response.Title)
			assert.Equal(t, "Test Content", response.Content)
			assert.Equal(t, nil, response.Price)
			assert.Equal(t, , response.AuthorId)
			assert.Equal(t, true, response.PublishedAt)
			assert.Equal(t, true, response.IsFeatured)
			assert.Equal(t, nil, response.Count)
			
		})

		t.Run("Update BlogPost", func(t *testing.T) {
			// Create test blogpost
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

			updateReq := models.UpdateBlogPostRequest{
				Title: nil,
				Content: "Updated Content",
				Price: nil,
				AuthorId: ,
				PublishedAt: false,
				IsFeatured: false,
				Count: nil,
				
			}

			jsonData, _ := json.Marshal(updateReq)
			req, _ := http.NewRequest("PUT", fmt.Sprintf("/api/blog-posts/%d", blogpost.Id), bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response models.BlogPostResponse
			err = json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, blogpost.Id, response.Id)
			assert.Equal(t, nil, response.Title)
			assert.Equal(t, "Updated Content", response.Content)
			assert.Equal(t, nil, response.Price)
			assert.Equal(t, , response.AuthorId)
			assert.Equal(t, false, response.PublishedAt)
			assert.Equal(t, false, response.IsFeatured)
			assert.Equal(t, nil, response.Count)
			
		})

		t.Run("Delete BlogPost", func(t *testing.T) {
			// Create test blogpost
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

			req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/blog-posts/%d", blogpost.Id), nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusNoContent, w.Code)

			// Verify deletion
			var found models.BlogPost
			err = helper.DB.First(&found, blogpost.Id).Error
			assert.Error(t, err) // Should not find deleted record
		})

		t.Run("List BlogPosts (paginated)", func(t *testing.T) {
			// Create multiple test blogposts
			for i := 0; i < 5; i++ {
				blogpost := &models.BlogPost{
					Title: nil,
					Content: fmt.Sprintf("Test Content %d", i),
					Price: nil,
					AuthorId: ,
					Author: ,
					PublishedAt: (i%2 == 0),
					IsFeatured: (i%2 == 0),
					Image: ,
					Count: nil,
					
				}
				err := helper.DB.Create(blogpost).Error
				assert.NoError(t, err)
			}

			req, _ := http.NewRequest("GET", "/api/blog-posts?page=1&limit=3", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response struct {
				Data       []models.BlogPostListResponse `json:"data"`
				Pagination struct {
					Page     int `json:"page"`
					PageSize int `json:"page_size"`
					Total    int `json:"total"`
				} `json:"pagination"`
			}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.NotEmpty(t, response.Data)
			assert.LessOrEqual(t, len(response.Data), 3)
			assert.Equal(t, 1, response.Pagination.Page)
			assert.Equal(t, 3, response.Pagination.PageSize)
		})

		t.Run("List All BlogPosts (unpaginated)", func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/api/blog-posts/all", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response []models.BlogPostListResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.NotNil(t, response)
		})

		t.Run("Error cases", func(t *testing.T) {
			// Test Get with invalid ID
			req, _ := http.NewRequest("GET", "/api/blog-posts/99999", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, http.StatusNotFound, w.Code)

			// Test Update with invalid ID
			updateReq := models.UpdateBlogPostRequest{
				Title: nil,
				Content: "Updated Content",
				Price: nil,
				AuthorId: ,
				PublishedAt: false,
				IsFeatured: false,
				Count: nil,
				
			}
			jsonData, _ := json.Marshal(updateReq)
			req, _ = http.NewRequest("PUT", "/api/blog-posts/99999", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			w = httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, http.StatusNotFound, w.Code)

			// Test Delete with invalid ID
			req, _ = http.NewRequest("DELETE", "/api/blog-posts/99999", nil)
			w = httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, http.StatusNotFound, w.Code)
		})
	})
}
