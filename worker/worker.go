package main

import (
	coderunner "code-scheduler/code-runner"
	"code-scheduler/db"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
)

var maxWorkers int
var activeWorkers int
var mu sync.Mutex
var workerName string
var configFile string

type Config struct {
    DB struct {
        URI      string `json:"uri"`
        Database string `json:"database"`
    } `json:"db"`
    WorkerName string `json:"worker_name"`
}
var config Config

func init() {
    flag.StringVar(&configFile, "config", "worker_config.json", "Path to configuration file")
}

func main() {
    flag.Parse()

    if len(os.Args) < 2 {
        log.Fatal("Usage: ./worker <max_workers> -config <config_file>")
    }

    maxWorkers, _ = strconv.Atoi(os.Args[1])
    loadConfig()

    err := db.Connect(config.DB.URI, config.DB.Database)
    if err != nil {
        log.Fatalf("Error connecting to database: %v", err)
    }

    http.HandleFunc("/process", processHandler)
    http.HandleFunc("/status", statusHandler)
    log.Fatal(http.ListenAndServe(":8082", nil))
}

func loadConfig() {
    data, err := os.ReadFile(configFile)
    if err != nil {
        log.Fatalf("Error reading config file: %v", err)
    }
    if err := json.Unmarshal(data, &config); err != nil {
        log.Fatalf("Error parsing config file: %v", err)
    }
    workerName = config.WorkerName
}
// Removed unused statusHandler function
func statusHandler(w http.ResponseWriter, r *http.Request) {
    response := map[string]int{
        "max_workers":    maxWorkers,
        "active_workers": activeWorkers,
    }
    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(response); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}
func processHandler(w http.ResponseWriter, r *http.Request) {
    jobID := r.URL.Query().Get("job_id")
    if jobID == "" {
        http.Error(w, "Missing job_id", http.StatusBadRequest)
        return
    }
    fmt.Println("TO PROCESS jobID: ", jobID)
    fmt.Println("activeWorkers: ", activeWorkers)
    mu.Lock()
    if activeWorkers >= maxWorkers {
        mu.Unlock()
        http.Error(w, "Max workers reached", http.StatusTooManyRequests)
        return
    }
    activeWorkers++
    mu.Unlock()

    go func() {
        defer func() {
            mu.Lock()
            activeWorkers--
            mu.Unlock()
        }()
        fmt.Println("marking job as scheduled")
        job, err := db.GetJob(jobID)

        // Fetch the job details from the database
        if err != nil {
            log.Printf("Error fetching job: %v", err)
            return
        }
        opts := db.ChangeStatusOptions{
                Status:     "Scheduled",
                WorkerName: workerName,
                Message:    "", // Empty message to trigger default
        }
        err = db.ChangeStatus(jobID, opts)
        if err != nil {
            log.Printf("Error updating job status scheduled: %v", err)
            return
        }
        
        // log the code running
        // Run the Python code from the job
        fmt.Println("Waiting for code to run to finish...")
        output, err := coderunner.RunPythonCode(job.PythonCode)
        if err != nil {
            log.Printf("Error running Python code: %v", err)
            // update status as errored and add a message
            opts := db.ChangeStatusOptions{
                Status:     "Errored",
                WorkerName: workerName,
                Message:    err.Error(), // Empty message to trigger default
            }
            err = db.ChangeStatus(jobID, opts)
            if err != nil {
                log.Printf("Error updating job status errored: %v", err)
            }
            return
        }

        // Create a file with the current timestamp in its name
     
      
        // Write the output of the Python code to the file
        fmt.Println("Marking as Finished")
        opts = db.ChangeStatusOptions{
                Status:     "Finished",
                WorkerName: workerName,
                Message:    output, // Empty message to trigger default
        }
        err = db.ChangeStatus(jobID, opts)
        if err != nil {
            log.Printf("Error updating job status Finished: %v", err)
            return
        }

        fmt.Fprintf(w, "Output: %s", output)
    }()
}