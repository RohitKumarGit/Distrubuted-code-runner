package db

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Job struct {
    ID         primitive.ObjectID `bson:"_id,omitempty"`
    PythonCode string             `bson:"python_code"`
    Status     string             `bson:"status"`
    WorkerName string             `bson:"worker_name,omitempty"`
    Message string               `bson:"message,omitempty"`
    CreatedAt  time.Time          `bson:"created_at"`
}

var client *mongo.Client
var jobsCollection *mongo.Collection

func Connect(uri, dbName string) error {
    var err error
    client, err = mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
    if err != nil {
        return err
    }
    jobsCollection = client.Database(dbName).Collection("jobs")
    return nil
}

func InsertJob(pythonCode string) (string, error) {
    job := Job{
        PythonCode: pythonCode,
        Status:     "In queue",
        CreatedAt:  time.Now(),
    }
    result, err := jobsCollection.InsertOne(context.TODO(), job)
    if err != nil {
        return "", err
    }
    return result.InsertedID.(primitive.ObjectID).Hex(), nil
}
type ChangeStatusOptions struct {
    Status     string
    WorkerName string
    Message    string
}
func    ChangeStatus(jobID string, opts ChangeStatusOptions) error {
     objID, err := primitive.ObjectIDFromHex(jobID)
    if err != nil {
        return err
    }

    // Set default message if not provided
    if opts.Message == "" {
        opts.Message = "No additional message provided."
    }

    filter := bson.M{"_id": objID}
    update := bson.M{
        "$set": bson.M{
            "status":      opts.Status,
            "worker_name": opts.WorkerName,
            "message":     opts.Message,
        },
    }
    _, err = jobsCollection.UpdateOne(context.TODO(), filter, update)
    return err
}

func GetJob(jobID string) (*Job, error) {
    objID, err := primitive.ObjectIDFromHex(jobID)
    if err != nil {
        return nil, err
    }
    filter := bson.M{"_id": objID}
    var job Job
    err = jobsCollection.FindOne(context.TODO(), filter).Decode(&job)
    if err != nil {
        return nil, err
    }
    return &job, nil
}

func GetQueuedJobs() ([]Job, error) {
    filter := bson.M{"status": "In queue"}
    cursor, err := jobsCollection.Find(context.TODO(), filter)
    if err != nil {
        return nil, err
    }
    var jobs []Job
    if err := cursor.All(context.TODO(), &jobs); err != nil {
        return nil, err
    }
    return jobs, nil
}