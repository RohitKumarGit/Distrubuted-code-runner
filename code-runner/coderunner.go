package coderunner

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
)

// RunPythonCode executes the given Python code and returns the output or an error.
func RunPythonCode(code string) (string, error) {
    // Create a temporary Python file
    tmpFile, err := ioutil.TempFile("", "code-*.py")
    
    if err != nil {
        return "", err
    }
    defer os.Remove(tmpFile.Name())

    // Write the Python code to the temporary file
    if _, err := tmpFile.Write([]byte(code)); err != nil {
        return "", err
    }
    if err := tmpFile.Close(); err != nil {
        return "", err
    }

    // Execute the Python script
    // run the python codde at the temporary file location
    // print location of the file
    cmd := exec.Command("python3", tmpFile.Name())
    var out bytes.Buffer
    var stderr bytes.Buffer
    cmd.Stdout = &out
    cmd.Stderr = &stderr
    fmt.Println("Running Python code...")
    fmt.Println("Python code output:")
    fmt.Println(cmd.Stdout)
    fmt.Println("Python code error:")
    fmt.Println(cmd.Stderr)
    if err := cmd.Run(); err != nil {
        return stderr.String(), err
    }

    return out.String(), nil
}