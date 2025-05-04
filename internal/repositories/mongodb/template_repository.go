package mongodb

import (
	"context"
	"time"

	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/models"
	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/repositories"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TemplateRepository implements the repositories.TemplateRepository interface
type TemplateRepository struct {
	collection *mongo.Collection
}

// NewTemplateRepository creates a new TemplateRepository
func NewTemplateRepository(db *mongo.Database) repositories.TemplateRepository {
	return &TemplateRepository{
		collection: db.Collection("templates"),
	}
}

// FindByID finds a template by ID
func (r *TemplateRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Template, error) {
	var template models.Template
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&template)
	if err != nil {
		return nil, err
	}
	return &template, nil
}

// FindByName finds a template by name
func (r *TemplateRepository) FindByName(ctx context.Context, name string) (*models.Template, error) {
	var template models.Template
	err := r.collection.FindOne(ctx, bson.M{"name": name}).Decode(&template)
	if err != nil {
		return nil, err
	}
	return &template, nil
}

// FindByType finds templates by type with pagination
func (r *TemplateRepository) FindByType(ctx context.Context, templateType string, page, limit int) ([]*models.Template, error) {
	opts := options.Find().
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit)).
		SetSort(bson.M{"name": 1}) // Sort by name ascending

	cursor, err := r.collection.Find(ctx, bson.M{"type": templateType}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var templates []*models.Template
	if err := cursor.All(ctx, &templates); err != nil {
		return nil, err
	}
	return templates, nil
}

// FindAll finds all templates with pagination
func (r *TemplateRepository) FindAll(ctx context.Context, page, limit int) ([]*models.Template, error) {
	opts := options.Find().
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit)).
		SetSort(bson.M{"name": 1}) // Sort by name ascending

	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var templates []*models.Template
	if err := cursor.All(ctx, &templates); err != nil {
		return nil, err
	}
	return templates, nil
}

// Create creates a new template
func (r *TemplateRepository) Create(ctx context.Context, template *models.Template) error {
	template.CreatedAt = time.Now()
	template.UpdatedAt = time.Now()
	_, err := r.collection.InsertOne(ctx, template)
	return err
}

// Update updates a template
func (r *TemplateRepository) Update(ctx context.Context, template *models.Template) error {
	template.UpdatedAt = time.Now()
	_, err := r.collection.ReplaceOne(ctx, bson.M{"_id": template.ID}, template)
	return err
}

// Delete deletes a template
func (r *TemplateRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

// Count counts all templates
func (r *TemplateRepository) Count(ctx context.Context) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{})
}
