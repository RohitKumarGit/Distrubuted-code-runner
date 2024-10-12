package main

import (
	"code-scheduler/db"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type Config struct {
    DB struct {
        URI      string `json:"uri"`
        Database string `json:"database"`
    } `json:"db"`
    Workers      []WorkerConfig `json:"workers"`
    CheckInterval string        `json:"check_interval"`
}

type WorkerConfig struct {
    IP   string `json:"ip"`
    Port int    `json:"port"`
}

var config Config

func main() {
    fmt.Println("Master started...")
    loadConfig()

    err := db.Connect(config.DB.URI, config.DB.Database)
    if err != nil {
        log.Fatalf("Error connecting to database: %v", err)
    }

    go startScheduler()

    http.HandleFunc("/submit", submitHandler)
    http.HandleFunc("/status", statusHandler)
    log.Fatal(http.ListenAndServe(":8080", nil))
}


func statusHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Println("Checking worker status...")

    type statusResponseWorker struct {
        MaxWorkers    int    `json:"max_workers"`
        ActiveWorkers int    `json:"active_workers"`
        WorkerName    string `json:"worker_name"`
    }

    type Response struct {
        Workers []statusResponseWorker `json:"workers"`
    }

    var response Response

    for _, worker := range config.Workers {
        url := fmt.Sprintf("http://%s:%d/status", worker.IP, worker.Port)
        resp, err := http.Get(url)
        if err != nil {
            log.Printf("Error fetching status from worker %s: %v", worker.IP, err)
            response.Workers = append(response.Workers, statusResponseWorker{
                MaxWorkers:    -1,
                ActiveWorkers: -1,
                WorkerName:    fmt.Sprintf("%s:%d", worker.IP, worker.Port),
            })
            continue
        }
        defer resp.Body.Close()

        if resp.StatusCode != http.StatusOK {
            log.Printf("Worker %s returned status %d", worker.IP, resp.StatusCode)
            response.Workers = append(response.Workers, statusResponseWorker{
                MaxWorkers:    -1,
                ActiveWorkers: -1,
                WorkerName:    fmt.Sprintf("%s:%d", worker.IP, worker.Port),
            })
            continue
        }

        var workerStatus statusResponseWorker
        if err := json.NewDecoder(resp.Body).Decode(&workerStatus); err != nil {
            log.Printf("Error decoding status from worker %s: %v", worker.IP, err)
            response.Workers = append(response.Workers, statusResponseWorker{
                MaxWorkers:    -1,
                ActiveWorkers: -1,
                WorkerName:    fmt.Sprintf("%s:%d", worker.IP, worker.Port),
            })
            continue
        }

        workerStatus.WorkerName = fmt.Sprintf("%s:%d", worker.IP, worker.Port)
        response.Workers = append(response.Workers, workerStatus)
    }

    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(response); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}


func loadConfig() {
    data, err := os.ReadFile("master_config.json")
    if err != nil {
        log.Fatalf("Error reading config file: %v", err)
    }
    if err := json.Unmarshal(data, &config); err != nil {
        log.Fatalf("Error parsing config file: %v", err)
    }
}
type SubmitRequest struct {
    PythonCode string `json:"python_code"`
}
func submitHandler(w http.ResponseWriter, r *http.Request) {
    // check python code in the request json body
    // if not found, return error
    var req SubmitRequest

    // Parse the JSON body
    decoder := json.NewDecoder(r.Body)
    err := decoder.Decode(&req)
    if err != nil {
        http.Error(w, "Invalid JSON body", http.StatusBadRequest)
        return
    }
    pythonCode := req.PythonCode
    if pythonCode == "" {
        http.Error(w, "Missing python_code", http.StatusBadRequest)
        return
    }

    jobID, err := db.InsertJob(pythonCode)
    if err != nil {
        log.Printf("Error inserting job: %v", err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    fmt.Fprintf(w, "Job submitted with ID: %s", jobID)
}

func startScheduler() {
    fmt.Println("Scheduler started...")
    interval, _ := time.ParseDuration(config.CheckInterval)
    ticker := time.NewTicker(interval)
    for range ticker.C {
        fmt.Println("Checking for queued jobs...")
        scheduleJobs()
    }
}

func scheduleJobs() {
    jobs, err := db.GetQueuedJobs()
    if err != nil {
        log.Printf("Error fetching queued jobs: %v", err)
        return
    }

    for _, job := range jobs {
        for _, worker := range config.Workers {
            fmt.Printf("Scheduling job %s to worker %s\n", job.ID, worker.IP)

            url := fmt.Sprintf("http://%s:%d/process?job_id=%s", worker.IP, worker.Port, job.ID.Hex())
            resp, err := http.Post(url, "application/json", nil)
            if err != nil {
                log.Printf("Error scheduling job %s to worker %s: %v", job.ID, worker.IP, err)
                continue
            }
            // print response status
            fmt.Println("Response status: ", resp.Status)
            if resp.StatusCode == http.StatusOK {
                log.Printf("Job %s scheduled to worker %s", job.ID, worker.IP)
                break
            }
        }
    }
}