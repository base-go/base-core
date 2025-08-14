package events_test

import (
	"testing"
	"time"

	"base/app/models"
	"base/core/translation"
	"base/core/types"
	"base/test"

	"github.com/stretchr/testify/assert"
)

func TestEventModels(t *testing.T) {
	helper := test.SetupTest(t)
	defer helper.TeardownTest()

	// Auto-migrate Event table for testing
	err := helper.DB.AutoMigrate(&models.Event{})
	assert.NoError(t, err)

	t.Run("Event model operations comprehensive coverage", func(t *testing.T) {
		t.Run("create and basic operations", func(t *testing.T) {
			// Test Event creation
			event := &models.Event{
				Title:            translation.NewField("Test Title"),
				Description:      translation.NewField("Test Description"),
				Location:         "Test Location",
				StartDate:        types.Now(),
				EndDate:          types.Now(),
				AllDay:           true,
				Recurring:        true,
				MaxAttendees:     123,
				CurrentAttendees: 123,
				Price:            123.45,
				Currency:         "Test Currency",
				Status:           "Test Status",
			}
			err := helper.DB.Create(event).Error
			assert.NoError(t, err)
			assert.NotZero(t, event.Id)
			assert.NotZero(t, event.CreatedAt)
			assert.NotZero(t, event.UpdatedAt)

			// Test find by ID
			var foundEvent models.Event
			err = helper.DB.First(&foundEvent, event.Id).Error
			assert.NoError(t, err)
			assert.Equal(t, event.Id, foundEvent.Id)
			assert.Equal(t, translation.NewField("Test Title"), foundEvent.Title)
			assert.Equal(t, translation.NewField("Test Description"), foundEvent.Description)
			assert.Equal(t, "Test Location", foundEvent.Location)
			assert.Equal(t, types.Now(), foundEvent.StartDate)
			assert.Equal(t, types.Now(), foundEvent.EndDate)
			assert.Equal(t, true, foundEvent.AllDay)
			assert.Equal(t, true, foundEvent.Recurring)
			assert.Equal(t, 123, foundEvent.MaxAttendees)
			assert.Equal(t, 123, foundEvent.CurrentAttendees)
			assert.Equal(t, 123.45, foundEvent.Price)
			assert.Equal(t, "Test Currency", foundEvent.Currency)
			assert.Equal(t, "Test Status", foundEvent.Status)

		})

		t.Run("Event model methods", func(t *testing.T) {
			event := &models.Event{
				Title:            translation.NewField("Test Title"),
				Description:      translation.NewField("Test Description"),
				Location:         "Test Location",
				StartDate:        types.Now(),
				EndDate:          types.Now(),
				AllDay:           true,
				Recurring:        true,
				MaxAttendees:     123,
				CurrentAttendees: 123,
				Price:            123.45,
				Currency:         "Test Currency",
				Status:           "Test Status",
			}
			err := helper.DB.Create(event).Error
			assert.NoError(t, err)

			// Test TableName
			assert.Equal(t, "events", event.TableName())

			// Test GetId
			assert.Equal(t, event.Id, event.GetId())

			// Test GetModelName
			assert.Equal(t, "event", event.GetModelName())

			// Test ToListResponse
			listResponse := event.ToListResponse()
			assert.NotNil(t, listResponse)
			assert.Equal(t, event.Id, listResponse.Id)
			assert.Equal(t, translation.NewField("Test Title"), listResponse.Title)
			assert.Equal(t, translation.NewField("Test Description"), listResponse.Description)
			assert.Equal(t, "Test Location", listResponse.Location)
			assert.Equal(t, types.Now(), listResponse.StartDate)
			assert.Equal(t, types.Now(), listResponse.EndDate)
			assert.Equal(t, true, listResponse.AllDay)
			assert.Equal(t, true, listResponse.Recurring)
			assert.Equal(t, 123, listResponse.MaxAttendees)
			assert.Equal(t, 123, listResponse.CurrentAttendees)
			assert.Equal(t, 123.45, listResponse.Price)
			assert.Equal(t, "Test Currency", listResponse.Currency)
			assert.Equal(t, "Test Status", listResponse.Status)

			// Test ToResponse
			response := event.ToResponse()
			assert.NotNil(t, response)
			assert.Equal(t, event.Id, response.Id)
			assert.Equal(t, translation.NewField("Test Title"), response.Title)
			assert.Equal(t, translation.NewField("Test Description"), response.Description)
			assert.Equal(t, "Test Location", response.Location)
			assert.Equal(t, types.Now(), response.StartDate)
			assert.Equal(t, types.Now(), response.EndDate)
			assert.Equal(t, true, response.AllDay)
			assert.Equal(t, true, response.Recurring)
			assert.Equal(t, 123, response.MaxAttendees)
			assert.Equal(t, 123, response.CurrentAttendees)
			assert.Equal(t, 123.45, response.Price)
			assert.Equal(t, "Test Currency", response.Currency)
			assert.Equal(t, "Test Status", response.Status)

			// Test Preload
			query := event.Preload(helper.DB)
			assert.NotNil(t, query)
		})

		t.Run("Event soft delete", func(t *testing.T) {
			event := &models.Event{
				Title:            translation.NewField("Test Title"),
				Description:      translation.NewField("Test Description"),
				Location:         "Test Location",
				StartDate:        types.Now(),
				EndDate:          types.Now(),
				AllDay:           true,
				Recurring:        true,
				MaxAttendees:     123,
				CurrentAttendees: 123,
				Price:            123.45,
				Currency:         "Test Currency",
				Status:           "Test Status",
			}
			err := helper.DB.Create(event).Error
			assert.NoError(t, err)

			// Test timestamps
			assert.True(t, event.CreatedAt.Before(time.Now().Add(time.Second)))
			assert.True(t, event.UpdatedAt.Before(time.Now().Add(time.Second)))

			// Test soft delete
			err = helper.DB.Delete(event).Error
			assert.NoError(t, err)

			// Should not find deleted record
			var foundEvent models.Event
			err = helper.DB.First(&foundEvent, event.Id).Error
			assert.Error(t, err)

			// Should find with Unscoped
			err = helper.DB.Unscoped().First(&foundEvent, event.Id).Error
			assert.NoError(t, err)
			assert.Equal(t, event.Id, foundEvent.Id)
			assert.NotZero(t, foundEvent.DeletedAt)
		})
	})
}
