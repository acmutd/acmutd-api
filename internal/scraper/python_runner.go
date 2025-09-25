package scraper

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

type PythonRunner struct {
	scraper string
}

func NewPythonRunner(scraper string) *PythonRunner {
	return &PythonRunner{
		scraper: scraper,
	}
}

/*
Executes the Python scraper script.
The scraper script is located in the scripts/{scraper}/main.py file.
If the venv/bin/activate file exists, it will use that to activate the virtual environment.
Otherwise, it will use the python command to run the script.
It will return the error if it fails.
*/
func (r *PythonRunner) Run() error {
	scraper := os.Getenv("SCRAPER")
	if scraper == "" {
		return fmt.Errorf("SCRAPER environment variable not set")
	}
	if _, err := os.Stat("scripts/" + scraper + "/main.py"); os.IsNotExist(err) {
		return fmt.Errorf("main.py not found in scripts/%s", scraper)
	}

	var cmd *exec.Cmd
	if _, err := os.Stat("venv/bin/activate"); !os.IsNotExist(err) {
		cmd = exec.Command("bash", "-c", "source ../../venv/bin/activate && python main.py")
	} else {
		cmd = exec.Command("python", "main.py")
	}
	cmd.Dir = filepath.Join("scripts", scraper)
	cmd.Stdout = log.Writer()
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "PYTHONPATH=/scripts/"+scraper)

	log.Println("Running Python scraper from directory:", cmd.Dir)
	log.Println("Python command:", cmd.String())

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run Python scraper: %w", err)
	}

	log.Println("Python scraper completed successfully")
	return nil
}
